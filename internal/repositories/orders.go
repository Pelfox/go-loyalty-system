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
}
