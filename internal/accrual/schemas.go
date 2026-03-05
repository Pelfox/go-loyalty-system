package accrual

// OrderStatus описывает текущее состояние заказа в системе начислений.
type OrderStatus string

const (
	// OrderStatusRegistered означает, что данный заказ был зарегистрирован в
	// системе, но начисление за него ещё не было посчитано.
	OrderStatusRegistered OrderStatus = "REGISTERED"
	// OrderStatusInvalid указывает на то, что заказ не был принят к расчёту, и
	// вознаграждение за него не будет засчитано.
	OrderStatusInvalid OrderStatus = "INVALID"
	// OrderStatusProcessing означает, что данный заказ был зарегистрирован, и
	// расчёт начисления в процессе.
	OrderStatusProcessing OrderStatus = "PROCESSING"
	// OrderStatusProcessed означает, что данный заказ был зарегистрирован и
	// начисление было рассчитано.
	OrderStatusProcessed OrderStatus = "PROCESSED"
)

// OrderResponse описывает ответ о заказе из внешней системы расчётов.
type OrderResponse struct {
	// Order - номер заказа, который данный ответ описывает.
	Order string `json:"order"`
	// Status - текущее состояние заказа.
	Status OrderStatus `json:"status"`
	// Accrual - начисление за данный заказ, если оно есть.
	Accrual *float64 `json:"accrual,omitempty"`
}
