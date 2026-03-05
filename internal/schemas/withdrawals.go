package schemas

import (
	"time"

	"github.com/Pelfox/go-loyalty-system/internal/models"
)

// UserBalanceResponse описывает структуру ответа на запрос о балансе
// пользователя.
type UserBalanceResponse struct {
	// Current - сумма, доступная к снятию для пользователя.
	Current float64 `json:"current"`
	// Withdrawn - сумма, которую пользователь снял за всё время.
	Withdrawn float64 `json:"withdrawn"`
}

// CreateWithdrawalRequest описывает запрос на списание вознаграждений с
// баланса пользователя.
type CreateWithdrawalRequest struct {
	// Order - номер заказа, с которого необходимо списать вознаграждения.
	Order string `json:"order" validate:"required"`
	// Sum - сумма для списания.
	Sum float64 `json:"sum" validate:"required"`
}

// WithdrawalResponse описывает ответ от сервера на запрос снятия
// вознаграждений со счёта пользователя.
type WithdrawalResponse struct {
	*models.Withdrawal
}

// WithdrawalItem описывает один уникальный элемент списания для пользователя.
type WithdrawalItem struct {
	// Order - номер заказа, к которому относится списание.
	Order string `json:"order"`
	// Sum - сумма данного списания.
	Sum float64 `json:"sum"`
	// ProcessedAt - дата и время списания.
	ProcessedAt time.Time `json:"processed_at"`
}
