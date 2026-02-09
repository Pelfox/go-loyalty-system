package repositories

import (
	"context"

	"github.com/Pelfox/go-loyalty-system/internal/models"
	"github.com/google/uuid"
)

// WithdrawalsRepository описывает возможные методы взаимодействия с сущностью
// транзакций списаний.
type WithdrawalsRepository interface {
	// Create создаёт новую транзакцию списания для указанных пользователя и
	// номера заказа.
	Create(ctx context.Context, userID uuid.UUID, orderNumber string, sum int64) (*models.Withdrawal, error)
	// GetForUser возвращает все транзакции списаний для пользователя с
	// указанным ID.
	GetForUser(ctx context.Context, userID uuid.UUID) ([]*models.Withdrawal, error)
	// GetUserBalance возвращает текущий баланс пользователя и общую сумму
	// списаний.
	GetUserBalance(ctx context.Context, userID uuid.UUID) (float64, int64, error)
}
