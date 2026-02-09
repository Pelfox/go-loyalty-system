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

func (r *OrdersRepository) UpdateStatus(
	ctx context.Context,
	number string,
	status models.OrderStatus,
	accrual *float64,
) error {
	query := "UPDATE orders SET status = $2, accrual = COALESCE($3, accrual) WHERE order_number = $1 AND status != 'PROCESSED'"
	_, err := r.db.Exec(ctx, query, number, status, accrual)
	return err
}
