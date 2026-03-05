package repositories

import (
	"context"

	"github.com/Pelfox/go-loyalty-system/internal/models"
	"github.com/google/uuid"
)

// OrdersRepository описывает возможные операции над models.Order.
type OrdersRepository interface {
	// Create создаёт новый заказ для указанного пользователя и с указанным номером
	// заказа. В случае удачного создания, возвращает объект models.Order с
	// созданным заказом, иначе - ошибку.
	//
	// Если один и тот же пользователь попробует создать второй заказ с одинаковым
	// ID, данная функция вернёт (nil, nil). В противном случае, в случае попытки
	// создания повторного заказа с одним номером, функция вернёт ошибку
	// repositories.ErrOrderAlreadyExists.
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
		accrual *float64,
	) error
}
