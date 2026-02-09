package postgres

import (
	"context"
	"errors"

	"github.com/Pelfox/go-loyalty-system/internal/models"
	"github.com/Pelfox/go-loyalty-system/internal/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrdersRepository реализует Postgres-репозиторий для
// repositories.OrdersRepository.
type OrdersRepository struct {
	db *pgxpool.Pool
}

// NewOrdersRepository создаёт и возвращает новый OrdersRepository.
func NewOrdersRepository(db *pgxpool.Pool) *OrdersRepository {
	return &OrdersRepository{db: db}
}

// Create создаёт новый заказ для указанного пользователя и с указанным номером
// заказа. В случае удачного создания, возвращает объект models.Order с
// созданным заказом, иначе - ошибку.
//
// Если один и тот же пользователь попробует создать второй заказ с одинаковым
// ID, данная функция вернёт (nil, nil). В противном случае, в случае попытки
// создания повторного заказа с одним номером, функция вернёт ошибку
// repositories.ErrOrderAlreadyExists.
func (r *OrdersRepository) Create(
	ctx context.Context,
	userID uuid.UUID,
	number string,
) (*models.Order, error) {
	var order models.Order
	order.UserID = userID
	order.Number = number

	query := "INSERT INTO orders(user_id, order_number) VALUES ($1, $2) RETURNING id, status, uploaded_at"
	err := r.db.QueryRow(ctx, query, userID, number).Scan(&order.ID, &order.Status, &order.UploadedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// название данного ограничения используется только тогда, когда
			// данный заказ был создан пользователем - это отдельный edge-case
			if pgErr.ConstraintName == "orders_same_user_unique" {
				return nil, nil
			}
			return nil, repositories.ErrOrderAlreadyExists
		}
		return nil, err
	}

	return &order, nil
}

// GetUserOrders возвращает все заказы для пользователя с указанным ID. Заказы
// отсортированы по времени создания от самых новых к самым старым. В случае
// ошибки данная функция возвращает её в неизменном виде.
func (r *OrdersRepository) GetUserOrders(
	ctx context.Context,
	userID uuid.UUID,
) ([]*models.Order, error) {
	query := "SELECT * FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC"
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*models.Order, 0)
	for rows.Next() {
		var order models.Order
		err = rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Accrual, &order.Status, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

// GetPendingOrders возвращает все заказы, которые нуждаются в обработке.
// Фильтр происходит по статусу заказа, значение которого должно быть или "NEW"
// или "PROCESSING". Заказы возвращаются от старых к новым.
func (r *OrdersRepository) GetPendingOrders(ctx context.Context, limit int) ([]*models.Order, error) {
	query := "SELECT * FROM orders WHERE status IN ('NEW', 'PROCESSING') ORDER BY uploaded_at LIMIT $1"
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*models.Order, 0)
	for rows.Next() {
		var order models.Order
		err = rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Accrual, &order.Status, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

// UpdateStatus обновляет статус заказа с указанным номером, устанавливая указанный
// новый статус. Опционально обновляет сумму начислений за заказ.
func (r *OrdersRepository) UpdateStatus(
	ctx context.Context,
	number string,
	status models.OrderStatus,
	accrual *int,
) error {
	query := "UPDATE orders SET status = $2, accrual = COALESCE($3, accrual) WHERE order_number = $1 AND status != 'PROCESSED'"
	_, err := r.db.Exec(ctx, query, number, status, accrual)
	return err
}
