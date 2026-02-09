package services

import (
	"context"
	"errors"

	"github.com/Pelfox/go-loyalty-system/internal/repositories"
	"github.com/Pelfox/go-loyalty-system/internal/schemas"
	"github.com/Pelfox/go-loyalty-system/pkg"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	// ErrInsufficientFunds описывает состояние, когда у пользователя
	// недостаточное количество доступных бонусов для снятия.
	ErrInsufficientFunds = errors.New("insufficient funds")
	// ErrWithdrawalCreateFailed используется тогда, когда сервис не может
	// списать вознаграждения со счёта пользователя.
	ErrWithdrawalCreateFailed = errors.New("failed to create withdrawal")
	// ErrWithdrawalsQueryFailed используется, если не удаётся получить все
	// списания пользователя.
	ErrWithdrawalsQueryFailed = errors.New("failed to query withdrawals")
)

// WithdrawalsService реализует всевозможные операции над сущностью снятия денег.
type WithdrawalsService struct {
	logger                zerolog.Logger
	withdrawalsRepository repositories.WithdrawalsRepository
}

// NewWithdrawalsService создаёт и возвращает новый объект WithdrawalsService.
func NewWithdrawalsService(
	logger zerolog.Logger,
	withdrawalsRepository repositories.WithdrawalsRepository,
) *WithdrawalsService {
	return &WithdrawalsService{
		logger:                logger.With().Str("service", "withdrawals").Logger(),
		withdrawalsRepository: withdrawalsRepository,
	}
}

// GetUserBalance получает и возвращает баланс пользователя, т.е. количество
// средств, доступных для снятия сейчас, и количество снятых средств в общем.
func (s *WithdrawalsService) GetUserBalance(
	ctx context.Context,
	userID uuid.UUID,
) (*schemas.UserBalanceResponse, error) {
	income, withdrawals, err := s.withdrawalsRepository.GetUserBalance(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("error getting user's income")
		return nil, err
	}

	return &schemas.UserBalanceResponse{
		Current:   income - float64(withdrawals),
		Withdrawn: withdrawals,
	}, nil
}

// CreateWithdrawal снимает указанную сумму со счёта вознаграждений указанного
// пользователя.
func (s *WithdrawalsService) CreateWithdrawal(
	ctx context.Context,
	userID uuid.UUID,
	orderNumber string,
	sum float64,
) (*schemas.WithdrawalResponse, error) {
	if !pkg.ValidateString(orderNumber) {
		return nil, ErrInvalidOrderNumber
	}

	withdrawal, err := s.withdrawalsRepository.Create(ctx, userID, orderNumber, sum)
	if err != nil {
		if errors.Is(err, repositories.ErrInsufficientFunds) {
			return nil, ErrInsufficientFunds
		}
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("error creating withdrawal")
		return nil, ErrWithdrawalCreateFailed
	}

	return &schemas.WithdrawalResponse{Withdrawal: withdrawal}, nil
}

// GetUserWithdrawals возвращает все операции снятия для данного пользователя.
func (s *WithdrawalsService) GetUserWithdrawals(ctx context.Context, userID uuid.UUID) ([]*schemas.WithdrawalItem, error) {
	withdrawals, err := s.withdrawalsRepository.GetForUser(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("user_id", userID.String()).Msg("error getting withdrawals")
		return nil, ErrWithdrawalsQueryFailed
	}

	items := make([]*schemas.WithdrawalItem, 0, len(withdrawals))
	for _, withdrawal := range withdrawals {
		items = append(items, &schemas.WithdrawalItem{
			Order:       withdrawal.OrderNumber,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt,
		})
	}

	return items, nil
}
