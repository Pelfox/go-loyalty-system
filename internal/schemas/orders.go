package schemas

import "github.com/Pelfox/go-loyalty-system/internal/models"

// CreateOrderResponse описывает формат ответа от сервера на запрос
// пользователя о создании нового заказа.
type CreateOrderResponse struct {
	*models.Order
}
