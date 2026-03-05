package controllers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Pelfox/go-loyalty-system/internal/constants"
	"github.com/Pelfox/go-loyalty-system/internal/services"
	"github.com/Pelfox/go-loyalty-system/pkg"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// OrdersController реализует обработчики запросов, относящихся к заказам.
type OrdersController struct {
	logger        zerolog.Logger
	ordersService *services.OrdersService
}

// NewOrdersController создаёт и возвращает новый OrdersController.
func NewOrdersController(
	logger zerolog.Logger,
	ordersService *services.OrdersService,
) *OrdersController {
	return &OrdersController{
		logger:        logger.With().Str("controller", "orders").Logger(),
		ordersService: ordersService,
	}
}

// ApplyRoutes регистрирует все пути, относящиеся к данному контроллеру.
func (c *OrdersController) ApplyRoutes(router chi.Router) {
	router.Post("/orders", c.Create)
	router.Get("/orders", c.GetUserOrders)
}

// Create сохраняет новый заказ по номеру, указанному пользователем.
func (c *OrdersController) Create(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		http.Error(w, "unexpected content type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	userID := r.Context().Value(constants.UserKey{}).(uuid.UUID)
	order, err := c.ordersService.Create(r.Context(), userID, string(body))
	if err != nil {
		if errors.Is(err, services.ErrOrderAlreadyExists) {
			http.Error(w, "order with this number already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, services.ErrInvalidOrderNumber) {
			http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// особый случай: пользователь пытается создать уже существующий заказ
	if order == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := pkg.WriteJSON(w, http.StatusAccepted, order); err != nil {
		c.logger.Error().Err(err).Msg("failed to write response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// GetUserOrders получает и возвращает все заказы пользователя.
func (c *OrdersController) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(constants.UserKey{}).(uuid.UUID)
	orders, err := c.ordersService.GetUserOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// если пользователь не имеет никаких заказов в БД
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := pkg.WriteJSON(w, http.StatusOK, orders); err != nil {
		c.logger.Error().Err(err).Msg("failed to write response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
