package schemas

import (
	"time"

	"github.com/Pelfox/go-loyalty-system/internal/models"
)

// CreateOrderResponse описывает формат ответа от сервера на запрос
// пользователя о создании нового заказа.
type CreateOrderResponse struct {
	*models.Order
}

// UserOrderItem описывает единый, уникальный заказ пользователя.
type UserOrderItem struct {
	// Number - номер этого заказа в системе.
	Number string `json:"number"`
	// Status - текущее состояние заказа.
	Status models.OrderStatus `json:"status"`
	// Accrual - количество начислений за данный заказ.
	Accrual *float64 `json:"accrual,omitempty"`
	// UploadedAt - дата и время, когда был добавлен заказ в систему.
	UploadedAt time.Time `json:"uploaded_at"`
}
