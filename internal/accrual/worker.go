package accrual

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Pelfox/go-loyalty-system/internal/models"
	"github.com/Pelfox/go-loyalty-system/internal/repositories"
	"github.com/rs/zerolog"
)

// workerQueryLimit - максимальное количество заказов, которое может обработать
// Worker за один batch.
const workerQueryLimit = 5

// Worker - основная абстракция для взаимодействия с внешней системой
// вознаграждений.
type Worker struct {
	logger       zerolog.Logger
	repository   repositories.OrdersRepository
	client       *Client
	pollInterval time.Duration

	pauseMu    sync.Mutex
	pauseUntil time.Time
	pauseChan  chan struct{}
}

// NewWorker создаёт и возвращает новый объект Worker, настраивая логгер из
// родительского, репозиторий и время между запросами к внешней системе.
func NewWorker(
	logger zerolog.Logger,
	repository repositories.OrdersRepository,
	client *Client,
	pollInterval time.Duration,
) *Worker {
	return &Worker{
		logger:       logger.With().Str("service", "accrual_worker").Logger(),
		repository:   repository,
		client:       client,
		pollInterval: pollInterval,
		pauseChan:    make(chan struct{}),
	}
}

// pause обновляет текущее состояние диспетчера воркеров, устанавливая время
// ожидания до следующей операции через duration.
func (w *Worker) pause(duration time.Duration) {
	until := time.Now().Add(duration)
	w.pauseMu.Lock()
	defer w.pauseMu.Unlock()

	if until.After(w.pauseUntil) {
		w.pauseUntil = until
		close(w.pauseChan) // будим всех, кто ждёт до истечения паузы
		w.pauseChan = make(chan struct{})
	}
}

// sleepIfPaused блокирует основной поток горутины, если один из воркеров
// обозначил необходимость ожидания до следующего запроса.
func (w *Worker) sleepIfPaused(ctx context.Context) error {
	for {
		w.pauseMu.Lock()
		until := w.pauseUntil
		pauseChan := w.pauseChan
		w.pauseMu.Unlock()

		now := time.Now()
		if !until.After(now) {
			return nil
		}

		timer := time.NewTimer(until.Sub(now))
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-pauseChan:
			// пересчитываем таймер, если кто-то продлил паузу
			timer.Stop()
			continue
		case <-timer.C:
			return nil
		}
	}
}

// Run запускает фоновый обработчик для внешней системы вознаграждений. Данная
// функция является блокирующей, поэтому её необходимо запускать как отдельную
// goroutine.
//
// Помимо этого, создаёт workersSize обработчиков поступающих запросов.
//
// Внутренне делает запросы в интервале pollInterval (из конструктора).
// Завершает свою работу вместе с завершением переданного context.Context.
func (w *Worker) Run(ctx context.Context, workersSize int) {
	jobs := make(chan models.Order, workerQueryLimit*2)
	var wg sync.WaitGroup

	// минимально ставим 1 воркер, дабы избежать дедлок
	if workersSize <= 0 {
		workersSize = 1
	}
	wg.Add(workersSize)

	// запускаем внутренние воркеры, обрабатывающие заказы от диспетчера
	for range workersSize {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}
					w.processSingle(ctx, job)
				}
			}
		}()
	}

	defer func() {
		close(jobs)
		wg.Wait()
		w.logger.Info().Msg("shutting down accrual worker")
	}()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// ожидаем необходимое количество времени, перед тем как продолжить
			// раздавать задачи
			if err := w.sleepIfPaused(ctx); err != nil && !errors.Is(err, context.Canceled) {
				w.logger.Error().Err(err).Msg("failed to sleep")
				return
			}

			orders, err := w.repository.GetPendingOrders(ctx, workerQueryLimit)
			if err != nil {
				w.logger.Error().Err(err).Msg("failed to get pending orders")
				continue
			}
			for _, order := range orders {
				select {
				case <-ctx.Done():
					return
				case jobs <- *order:
				}
			}
		}
	}
}

// processSingle обрабатывает уникальный и единственный заказ, полученный от
// вышестоящего диспетчера операций.
func (w *Worker) processSingle(ctx context.Context, order models.Order) {
	// ожидаем необходимое количество времени перед совершением запроса
	if err := w.sleepIfPaused(ctx); err != nil {
		w.logger.Error().Err(err).Msg("failed to sleep before processing the order")
		return
	}

	w.logger.Debug().Str("order_number", order.Number).Msg("processing order")
	resp, err := w.client.GetOrder(ctx, order.Number)
	if err != nil {
		// заказ ещё не появился во внешнем сервисе
		if errors.Is(err, ErrNotRegistered) {
			return
		}
		var rateLimitError RateLimitError
		// попали в лимиты
		if ok := errors.As(err, &rateLimitError); ok {
			w.logger.Warn().
				Dur("retry_after", rateLimitError.RetryAfter).
				Msg("rate limit exceeded, retrying")

			// ставим глобальную паузу для воркеров до следующего запроса
			w.pause(rateLimitError.RetryAfter)
			return
		}
		w.logger.Error().Err(err).Str("number", order.Number).Msg("failed to fetch the order")
		return
	}

	switch resp.Status {
	case OrderStatusRegistered, OrderStatusProcessing:
		if err := w.repository.UpdateStatus(ctx, resp.Order, models.OrderStatusProcessing, nil); err != nil {
			w.logger.Error().Err(err).
				Str("order_id", order.ID.String()).
				Str("new_status", "processing").
				Msg("failed to update status for order")
		}
	case OrderStatusInvalid:
		if err := w.repository.UpdateStatus(ctx, resp.Order, models.OrderStatusInvalid, nil); err != nil {
			w.logger.Error().Err(err).
				Str("order_id", order.ID.String()).
				Str("new_status", "invalid").
				Msg("failed to update status for order")
		}
	case OrderStatusProcessed:
		if err := w.repository.UpdateStatus(ctx, resp.Order, models.OrderStatusProcessed, resp.Accrual); err != nil {
			w.logger.Error().Err(err).
				Str("order_id", order.ID.String()).
				Str("new_status", "processed").
				Msg("failed to update status for order")
		}
	}
}
