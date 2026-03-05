package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Pelfox/go-loyalty-system/internal/models"
	"github.com/Pelfox/go-loyalty-system/internal/repositories"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// fakeOrdersRepository - тестовая реализация OrdersRepository для
// тестирования OrdersService.
type fakeOrdersRepository struct {
	createFunc        func(ctx context.Context, userID uuid.UUID, orderNumber string) (*models.Order, error)
	getUserOrdersFunc func(ctx context.Context, userID uuid.UUID) ([]*models.Order, error)
}

func (f *fakeOrdersRepository) GetPendingOrders(_ context.Context, _ int) ([]*models.Order, error) {
	panic("this method is not supported by the mocking repository - do not call it")
}

func (f *fakeOrdersRepository) UpdateStatus(_ context.Context, _ string, _ models.OrderStatus, _ *float64) error {
	panic("this method is not supported by the mocking repository - do not call it")
}

func (f *fakeOrdersRepository) Create(
	ctx context.Context,
	userID uuid.UUID,
	orderNumber string,
) (*models.Order, error) {
	return f.createFunc(ctx, userID, orderNumber)
}

func (f *fakeOrdersRepository) GetUserOrders(
	ctx context.Context,
	userID uuid.UUID,
) ([]*models.Order, error) {
	return f.getUserOrdersFunc(ctx, userID)
}

// TestOrdersService_Create_OK проверяет успешное создание нового заказа.
func TestOrdersService_Create_OK(t *testing.T) {
	userID := uuid.New()
	orderNumber := "12345678903" // валидный номер по Луну

	repo := &fakeOrdersRepository{
		createFunc: func(ctx context.Context, uid uuid.UUID, number string) (*models.Order, error) {
			return &models.Order{
				ID:     uuid.New(),
				UserID: uid,
				Number: number,
			}, nil
		},
	}
	service := NewOrdersService(
		zerolog.Nop(),
		repo,
	)

	resp, err := service.Create(context.Background(), userID, orderNumber)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil || resp.Order == nil {
		t.Fatalf("response or order is nil")
	}

	if resp.Order.UserID != userID {
		t.Errorf("userID = %v, want %v", resp.Order.UserID, userID)
	}

	if resp.Order.Number != orderNumber {
		t.Errorf("order number = %q, want %q", resp.Order.Number, orderNumber)
	}
}

// TestOrdersService_Create_InvalidOrderNumber проверяет действие сервиса при
// некорректно-введённом номере заказа.
func TestOrdersService_Create_InvalidOrderNumber(t *testing.T) {
	repo := &fakeOrdersRepository{}
	service := NewOrdersService(
		zerolog.Nop(),
		repo,
	)

	_, err := service.Create(
		context.Background(),
		uuid.New(),
		"123456789", // невалидный номер
	)
	if !errors.Is(err, ErrInvalidOrderNumber) {
		t.Errorf("error = %v, want %v", err, ErrInvalidOrderNumber)
	}
}

// TestOrdersService_Create_OrderAlreadyExists тестирует сервис на повторное
// создание заказа с уже существующем номером заказа, привязанном к другому
// пользователю.
func TestOrdersService_Create_OrderAlreadyExists(t *testing.T) {
	repo := &fakeOrdersRepository{
		createFunc: func(ctx context.Context, userID uuid.UUID, orderNumber string) (*models.Order, error) {
			return nil, repositories.ErrOrderAlreadyExists
		},
	}
	service := NewOrdersService(
		zerolog.Nop(),
		repo,
	)

	_, err := service.Create(
		context.Background(),
		uuid.New(),
		"12345678903",
	)
	if !errors.Is(err, ErrOrderAlreadyExists) {
		t.Errorf("error = %v, want %v", err, ErrOrderAlreadyExists)
	}
}

// TestOrdersService_Create_RepositoryError тестирует сервис на обработку
// ошибок дочернего слоя (репозиторий).
func TestOrdersService_Create_RepositoryError(t *testing.T) {
	repo := &fakeOrdersRepository{
		createFunc: func(ctx context.Context, userID uuid.UUID, orderNumber string) (*models.Order, error) {
			return nil, errors.New("db error")
		},
	}
	service := NewOrdersService(
		zerolog.Nop(),
		repo,
	)

	_, err := service.Create(
		context.Background(),
		uuid.New(),
		"12345678903",
	)
	if !errors.Is(err, ErrOrderCreationFailed) {
		t.Errorf("error = %v, want %v", err, ErrOrderCreationFailed)
	}
}

// TestOrdersService_Create_OrderAlreadyExistsForUser тестирует поведение
// сервиса на специальном случае, когда данный заказ уже существует для этого
// пользователя.
func TestOrdersService_Create_OrderAlreadyExistsForUser(t *testing.T) {
	repo := &fakeOrdersRepository{
		createFunc: func(ctx context.Context, userID uuid.UUID, orderNumber string) (*models.Order, error) {
			return nil, nil
		},
	}

	service := NewOrdersService(
		zerolog.Nop(),
		repo,
	)
	resp, err := service.Create(
		context.Background(),
		uuid.New(),
		"12345678903",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp != nil {
		t.Errorf("response = %v, want nil", resp)
	}
}

// TestOrdersService_GetUserOrders_OK тестирует сервис на возвращение всех
// заказов пользователя.
func TestOrdersService_GetUserOrders_OK(t *testing.T) {
	userID := uuid.New()
	expected := []*models.Order{
		{
			ID:     uuid.New(),
			UserID: userID,
			Number: "12345678903",
		},
	}

	repo := &fakeOrdersRepository{
		getUserOrdersFunc: func(ctx context.Context, uid uuid.UUID) ([]*models.Order, error) {
			return expected, nil
		},
	}

	service := NewOrdersService(
		zerolog.Nop(),
		repo,
	)

	orders, err := service.GetUserOrders(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(orders) != len(expected) {
		t.Fatalf("orders len = %d, want %d", len(orders), len(expected))
	}

	if orders[0].Number != expected[0].Number {
		t.Errorf("order number = %q, want %q", orders[0].Number, expected[0].Number)
	}
}

// TestOrdersService_GetUserOrders_RepositoryError тестирует поведение сервиса,
// когда дочерний слой (репозиторий) возвращает ошибку при получении всех
// заказов пользователя.
func TestOrdersService_GetUserOrders_RepositoryError(t *testing.T) {
	repo := &fakeOrdersRepository{
		getUserOrdersFunc: func(ctx context.Context, userID uuid.UUID) ([]*models.Order, error) {
			return nil, errors.New("db error")
		},
	}

	service := NewOrdersService(
		zerolog.Nop(),
		repo,
	)

	_, err := service.GetUserOrders(context.Background(), uuid.New())
	if !errors.Is(err, ErrOrdersQueryFailed) {
		t.Errorf("error = %v, want %v", err, ErrOrdersQueryFailed)
	}
}
