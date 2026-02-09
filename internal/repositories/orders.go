package repositories

import (
	"context"

	"github.com/Pelfox/go-loyalty-system/internal/models"
	"github.com/google/uuid"
)

// OrdersRepository описывает возможные операции над models.Order.
type OrdersRepository interface {
	// Create создаёт новый заказ, привязывая его к пользователю с указанным ID
	// и задаёт указанный номер заказа.
	Create(ctx context.Context, userID uuid.UUID, number string) (*models.Order, error)
	// GetUserOrders запрашивает и возвращает все заказы, оформленные данным
	// пользователем. Заказы отсортированы от самых новых к самым старым.
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]*models.Order, error)

	// GetPendingOrders возвращает все заказы, которые необходимо обработать.
	GetPendingOrders(ctx context.Context, limit int) ([]*models.Order, error)
	// UpdateStatus обновляет статус заказа для заказа с указанным номером.
	// Опционально обновляет сумму баллов за данный заказ.
	UpdateStatus(
		ctx context.Context,
		orderNumber string,
		status models.OrderStatus,
		accrual *int,
	) error
}
