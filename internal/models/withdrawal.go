package models

import (
	"time"

	"github.com/google/uuid"
)

// Withdrawal описывает уникальную транзакцию снятия денег с баланса за заказ.
type Withdrawal struct {
	// ID - уникальный идентификатор транзакции.
	ID uuid.UUID `json:"id"`
	// UserID - уникальный идентификатор пользователя, совершившего транзакцию.
	UserID uuid.UUID `json:"user_id"`
	// OrderNumber - номер заказа в системе, за который было списание.
	OrderNumber string `json:"order_number"`
	// Sum - сумма списания.
	Sum float64 `json:"sum"`
	// ProcessedAt - дата и время, когда было произведено списание.
	ProcessedAt time.Time `json:"processed_at"`
}
