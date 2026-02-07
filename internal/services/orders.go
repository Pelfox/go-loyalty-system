package services

import (
	"context"
	"errors"
	"strings"

	"github.com/Pelfox/go-loyalty-system/internal/models"
	"github.com/Pelfox/go-loyalty-system/internal/repositories"
	"github.com/Pelfox/go-loyalty-system/internal/schemas"
	"github.com/Pelfox/go-loyalty-system/pkg"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	// ErrOrderAlreadyExists используется когда пользователь пытается создать
	// заказ с номером, который уже существует, и не принадлежит ему.
	ErrOrderAlreadyExists = errors.New("order already exists")
	// ErrOrderCreationFailed обозначает, что не удаётся создать заказ.
	ErrOrderCreationFailed = errors.New("failed to create order")
	// ErrInvalidOrderNumber используется, если указанный номер заказа не
	// является корректным (проверка алгоритмом Луна).
	ErrInvalidOrderNumber = errors.New("invalid order number")
	// ErrOrdersQueryFailed используется, когда не удаётся получить заказы
	// пользователя из-за ошибки БД.
	ErrOrdersQueryFailed = errors.New("failed to query user's orders")
)

// OrdersService реализует взаимодействие с заказами в системе.
type OrdersService struct {
	logger           zerolog.Logger
	ordersRepository repositories.OrdersRepository
}

// NewOrdersService создаёт и возвращает OrdersService.
func NewOrdersService(
	logger zerolog.Logger,
	ordersRepository repositories.OrdersRepository,
) *OrdersService {
	return &OrdersService{
		logger:           logger.With().Str("service", "orders").Logger(),
		ordersRepository: ordersRepository,
	}
}

// Create создаёт новый заказ для пользователя с указанным ID и указанным
// номером заказа. Возвращает (nil, nil), если пользователь пытается создать
// уже существующий заказ, привязанный к его ID.
func (s *OrdersService) Create(
	ctx context.Context,
	userID uuid.UUID,
	orderNumber string,
) (*schemas.CreateOrderResponse, error) {
	orderNumber = strings.ReplaceAll(strings.TrimSpace(orderNumber), " ", "")

	// проверяем номер заказа через алгоритм Луна
	if !pkg.ValidateString(orderNumber) {
		return nil, ErrInvalidOrderNumber
	}

	order, err := s.ordersRepository.Create(ctx, userID, orderNumber)
	if err != nil {
		if errors.Is(err, repositories.ErrOrderAlreadyExists) {
			return nil, ErrOrderAlreadyExists
		}
		s.logger.Error().Err(err).Msg("failed to create an order")
		return nil, ErrOrderCreationFailed
	}

	// особый случай: пользователь пытается создать уже свой существующий заказ
	if order == nil {
		return nil, nil
	}

	return &schemas.CreateOrderResponse{Order: order}, nil
}

// GetUserOrders получает и возвращает все заказы пользователя с указанным ID.
func (s *OrdersService) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*models.Order, error) {
	orders, err := s.ordersRepository.GetUserOrders(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get orders for user")
		return nil, ErrOrdersQueryFailed
	}
	return orders, nil
}
