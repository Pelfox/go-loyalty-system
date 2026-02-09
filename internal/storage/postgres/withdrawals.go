package postgres

import (
	"context"

	"github.com/Pelfox/go-loyalty-system/internal/models"
	"github.com/Pelfox/go-loyalty-system/internal/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithdrawalsRepository реализовывает repositories.WithdrawalsRepository для
// PostgreSQL.
type WithdrawalsRepository struct {
	db *pgxpool.Pool
}

// NewWithdrawalsRepository создаёт и возвращает новый WithdrawalsRepository.
func NewWithdrawalsRepository(db *pgxpool.Pool) *WithdrawalsRepository {
	return &WithdrawalsRepository{db: db}
}

func (r *WithdrawalsRepository) Create(ctx context.Context, userID uuid.UUID, orderNumber string, sum float64) (*models.Withdrawal, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)
	var totalIncome float64
	var totalWithdrawn float64

	incomeQuery := "SELECT COALESCE(SUM(accrual), 0.0) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'"
	if err := tx.QueryRow(ctx, incomeQuery, userID).Scan(&totalIncome); err != nil {
		return nil, err
	}

	withdrawalsQuery := "SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1"
	if err := tx.QueryRow(ctx, withdrawalsQuery, userID).Scan(&totalWithdrawn); err != nil {
		return nil, err
	}

	if totalIncome-totalWithdrawn < sum {
		return nil, repositories.ErrInsufficientFunds
	}

	withdrawal := models.Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		Sum:         sum,
	}

	query := "INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3) RETURNING id, processed_at"
	err = tx.QueryRow(ctx, query, userID, orderNumber, sum).Scan(&withdrawal.ID, &withdrawal.ProcessedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &withdrawal, nil
}

func (r *WithdrawalsRepository) GetForUser(ctx context.Context, userID uuid.UUID) ([]*models.Withdrawal, error) {
	query := "SELECT * FROM withdrawals WHERE user_id = $1"
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals := make([]*models.Withdrawal, 0)
	for rows.Next() {
		var withdrawal models.Withdrawal
		err = rows.Scan(&withdrawal.ID, &withdrawal.UserID, &withdrawal.OrderNumber, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &withdrawal)
	}

	return withdrawals, nil
}

func (r *WithdrawalsRepository) GetUserBalance(ctx context.Context, userID uuid.UUID) (float64, float64, error) {
	var totalIncome float64
	var totalWithdrawn float64

	query := "SELECT COALESCE(SUM(accrual), 0.0) FROM orders WHERE user_id = $1 AND status = 'PROCESSED'"
	if err := r.db.QueryRow(ctx, query, userID).Scan(&totalIncome); err != nil {
		return 0, 0, err
	}

	withdrawalsQuery := "SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1"
	if err := r.db.QueryRow(ctx, withdrawalsQuery, userID).Scan(&totalWithdrawn); err != nil {
		return 0, 0, err
	}

	return totalIncome, totalWithdrawn, nil
}
