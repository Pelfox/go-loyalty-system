package accrual

import (
	"context"
	"errors"
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
	}
}

// Run запускает фоновый обработчик для внешней системы вознаграждений. Данная
// функция является блокирующей, поэтому её необходимо запускать как отдельную
// goroutine.
//
// Внутренне делает запросы в интервале pollInterval (из конструктора).
// Завершает свою работу вместе с завершением переданного context.Context.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info().Msg("shutting down accrual worker")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

// processBatch запрашивает все заказы, которые необходимо обработать, и
// пытается обработать их, используя Client для получения текущего статуса
// заказа во внешней системе вознаграждений.
func (w *Worker) processBatch(ctx context.Context) {
	orders, err := w.repository.GetPendingOrders(ctx, workerQueryLimit)
	if err != nil {
		w.logger.Error().Err(err).Msg("failed to get pending orders")
		return
	}

	for _, order := range orders {
		w.logger.Debug().Str("order_number", order.Number).Msg("processing order")
		resp, err := w.client.GetOrder(ctx, order.Number)
		if err != nil {
			// заказ ещё не появился во внешнем сервисе
			if errors.Is(err, ErrNotRegistered) {
				continue
			}
			var rateLimitError RateLimitError
			// попали в "лимиты"
			if ok := errors.As(err, &rateLimitError); ok {
				w.logger.Warn().
					Dur("retry_after", rateLimitError.RetryAfter).
					Msg("rate limit exceeded, retrying")
				time.Sleep(rateLimitError.RetryAfter)
				continue
			}
			w.logger.Error().Err(err).Str("number", order.Number).Msg("failed to fetch the order")
			continue
		}

		switch resp.Status {
		case OrderStatusRegistered, OrderStatusProcessing:
			_ = w.repository.UpdateStatus(ctx, resp.Order, models.OrderStatusProcessing, nil)
		case OrderStatusInvalid:
			_ = w.repository.UpdateStatus(ctx, resp.Order, models.OrderStatusInvalid, nil)
		case OrderStatusProcessed:
			_ = w.repository.UpdateStatus(ctx, resp.Order, models.OrderStatusProcessed, resp.Accrual)
		}
	}
}
