package models

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus описывает текущее состояние (статус) заказа.
type OrderStatus string

const (
	// OrderStatusNew используется, когда заказ был зарегистрирован, но ещё не
	// был обработан системой.
	OrderStatusNew OrderStatus = "NEW"
	// OrderStatusProcessing описывает состояние заказа, когда он находится в
	// процессе обработки.
	OrderStatusProcessing OrderStatus = "PROCESSING"
	// OrderStatusProcessed описывает состояние заказа, когда он уже был
	// обработан системой.
	OrderStatusProcessed OrderStatus = "PROCESSED"
	// OrderStatusInvalid указывает на то, что данный заказ не удалось
	// обработать.
	OrderStatusInvalid OrderStatus = "INVALID"
)

// Order описывает сущность одного заказа в системе.
type Order struct {
	// ID - уникальный идентификатор данного заказа.
	ID uuid.UUID `json:"id"`
	// UserID - идентификатор пользователя, создавшего заказ.
	UserID uuid.UUID `json:"user_id"`
	// Number - уникальный номер заказа.
	Number string `json:"number"`
	// Accrual - количество баллов, начисленное пользователю за заказ.
	Accrual *float64 `json:"accrual,omitempty"`
	// Status - текущее состояние заказа.
	Status OrderStatus `json:"status"`
	// UploadedAt - дата и время, когда заказ был создан.
	UploadedAt time.Time `json:"uploaded_at"`
}
