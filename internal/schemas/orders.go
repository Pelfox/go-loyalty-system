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

type UserOrderItem struct {
	Number     string             `json:"number"`
	Status     models.OrderStatus `json:"status"`
	Accrual    *int               `json:"accrual,omitempty"`
	UploadedAt time.Time          `json:"uploaded_at"`
}
