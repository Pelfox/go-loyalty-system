package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Pelfox/go-loyalty-system/internal/models"
	"github.com/Pelfox/go-loyalty-system/internal/repositories"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// fakeWithdrawalsRepository - mock-репозиторий для тестов.
type fakeWithdrawalsRepository struct {
	getUserBalanceFunc func(ctx context.Context, userID uuid.UUID) (float64, float64, error)
	createFunc         func(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) (*models.Withdrawal, error)
	getForUserFunc     func(ctx context.Context, userID uuid.UUID) ([]*models.Withdrawal, error)
}

func (f *fakeWithdrawalsRepository) GetUserBalance(
	ctx context.Context,
	userID uuid.UUID,
) (float64, float64, error) {
	return f.getUserBalanceFunc(ctx, userID)
}

func (f *fakeWithdrawalsRepository) Create(
	ctx context.Context,
	userID uuid.UUID,
	orderNumber string,
	sum float64,
) (*models.Withdrawal, error) {
	return f.createFunc(ctx, userID, orderNumber, sum)
}

func (f *fakeWithdrawalsRepository) GetForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]*models.Withdrawal, error) {
	return f.getForUserFunc(ctx, userID)
}

// TestWithdrawalsService_GetUserBalance_OK тестирует нормальное поведение
// сервиса.
func TestWithdrawalsService_GetUserBalance_OK(t *testing.T) {
	repo := &fakeWithdrawalsRepository{
		getUserBalanceFunc: func(ctx context.Context, userID uuid.UUID) (float64, float64, error) {
			return 1000, 300, nil
		},
	}

	service := NewWithdrawalsService(zerolog.Nop(), repo)
	resp, err := service.GetUserBalance(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Current != 700 {
		t.Errorf("current = %v, want %v", resp.Current, 700)
	}

	if resp.Withdrawn != 300 {
		t.Errorf("withdrawn = %v, want %v", resp.Withdrawn, 300)
	}
}

// TestWithdrawalsService_GetUserBalance_RepositoryError тестирует поведение
// сервиса при ошибке репозитория.
func TestWithdrawalsService_GetUserBalance_RepositoryError(t *testing.T) {
	repo := &fakeWithdrawalsRepository{
		getUserBalanceFunc: func(ctx context.Context, userID uuid.UUID) (float64, float64, error) {
			return 0, 0, errors.New("db error")
		},
	}

	service := NewWithdrawalsService(zerolog.Nop(), repo)
	_, err := service.GetUserBalance(context.Background(), uuid.New())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// TestWithdrawalsService_CreateWithdrawal_InvalidOrderNumber тестирует
// поведение репозитория при некорректном номере заказа.
func TestWithdrawalsService_CreateWithdrawal_InvalidOrderNumber(t *testing.T) {
	repo := &fakeWithdrawalsRepository{}
	service := NewWithdrawalsService(zerolog.Nop(), repo)
	_, err := service.CreateWithdrawal(context.Background(), uuid.New(), "", 100)
	if !errors.Is(err, ErrInvalidOrderNumber) {
		t.Errorf("error = %v, want %v", err, ErrInvalidOrderNumber)
	}
}

// TestWithdrawalsService_CreateWithdrawal_InsufficientFunds тестирует
// поведение сервиса при недостаточном количестве средств у пользователя.
func TestWithdrawalsService_CreateWithdrawal_InsufficientFunds(t *testing.T) {
	repo := &fakeWithdrawalsRepository{
		createFunc: func(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) (*models.Withdrawal, error) {
			return nil, repositories.ErrInsufficientFunds
		},
	}
	service := NewWithdrawalsService(zerolog.Nop(), repo)
	_, err := service.CreateWithdrawal(context.Background(), uuid.New(), "378282246310005", 100)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("error = %v, want %v", err, ErrInsufficientFunds)
	}
}

// TestWithdrawalsService_CreateWithdrawal_RepositoryError тестирует поведение
// сервиса при ошибке репозитория.
func TestWithdrawalsService_CreateWithdrawal_RepositoryError(t *testing.T) {
	repo := &fakeWithdrawalsRepository{
		createFunc: func(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) (*models.Withdrawal, error) {
			return nil, errors.New("db error")
		},
	}
	service := NewWithdrawalsService(zerolog.Nop(), repo)
	_, err := service.CreateWithdrawal(context.Background(), uuid.New(), "378282246310005", 100)
	if !errors.Is(err, ErrWithdrawalCreateFailed) {
		t.Errorf("error = %v, want %v", err, ErrWithdrawalCreateFailed)
	}
}

// TestWithdrawalsService_CreateWithdrawal_OK тестирует корректное поведение
// сервиса при создании нового снятия.
func TestWithdrawalsService_CreateWithdrawal_OK(t *testing.T) {
	now := time.Now()
	repo := &fakeWithdrawalsRepository{
		createFunc: func(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) (*models.Withdrawal, error) {
			return &models.Withdrawal{
				OrderNumber: orderNumber,
				Sum:         sum,
				ProcessedAt: now,
			}, nil
		},
	}

	service := NewWithdrawalsService(zerolog.Nop(), repo)
	resp, err := service.CreateWithdrawal(context.Background(), uuid.New(), "378282246310005", 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Withdrawal.Sum != 500 {
		t.Errorf("sum = %v, want %v", resp.Withdrawal.Sum, 500)
	}
}

// TestWithdrawalsService_GetUserWithdrawals_OK тестирует корректное поведение
// сервиса для получения всех снятий пользователя.
func TestWithdrawalsService_GetUserWithdrawals_OK(t *testing.T) {
	now := time.Now()
	repo := &fakeWithdrawalsRepository{
		getForUserFunc: func(ctx context.Context, userID uuid.UUID) ([]*models.Withdrawal, error) {
			return []*models.Withdrawal{
				{
					OrderNumber: "123",
					Sum:         100,
					ProcessedAt: now,
				},
			}, nil
		},
	}

	service := NewWithdrawalsService(zerolog.Nop(), repo)
	items, err := service.GetUserWithdrawals(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("len = %d, want 1", len(items))
	}

	if items[0].Order != "123" {
		t.Errorf("order = %v, want %v", items[0].Order, "123")
	}
}

// TestWithdrawalsService_GetUserWithdrawals_RepositoryError тестирует
// поведение сервиса при ошибке репозитория при получении всех снятий.
func TestWithdrawalsService_GetUserWithdrawals_RepositoryError(t *testing.T) {
	repo := &fakeWithdrawalsRepository{
		getForUserFunc: func(ctx context.Context, userID uuid.UUID) ([]*models.Withdrawal, error) {
			return nil, errors.New("db error")
		},
	}
	service := NewWithdrawalsService(zerolog.Nop(), repo)
	_, err := service.GetUserWithdrawals(context.Background(), uuid.New())
	if !errors.Is(err, ErrWithdrawalsQueryFailed) {
		t.Errorf("error = %v, want %v", err, ErrWithdrawalsQueryFailed)
	}
}
