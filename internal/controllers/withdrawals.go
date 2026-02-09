package controllers

import (
	"errors"
	"net/http"

	"github.com/Pelfox/go-loyalty-system/internal"
	"github.com/Pelfox/go-loyalty-system/internal/schemas"
	"github.com/Pelfox/go-loyalty-system/internal/services"
	"github.com/Pelfox/go-loyalty-system/pkg"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// WithdrawalsController реализовывает взаимодействие с сервисом снятия.
type WithdrawalsController struct {
	logger             zerolog.Logger
	withdrawalsService *services.WithdrawalsService
}

// NewWithdrawalsController создаёт и возвращает новый WithdrawalsController.
func NewWithdrawalsController(
	logger zerolog.Logger,
	withdrawalsService *services.WithdrawalsService,
) *WithdrawalsController {
	return &WithdrawalsController{
		logger:             logger.With().Str("controller", "withdrawals").Logger(),
		withdrawalsService: withdrawalsService,
	}
}

// ApplyRoutes применяет все пути данного обработчика к указанному роутеру.
func (c *WithdrawalsController) ApplyRoutes(router chi.Router) {
	router.Get("/balance", c.GetUserBalance)
	router.Post("/balance/withdraw", c.CreateWithdrawal)
	router.Get("/balance/withdraws", c.GetUserWithdrawals)
}

// GetUserBalance возвращает информацию о текущем состоянии счёта пользователя.
func (c *WithdrawalsController) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(internal.UserKey{}).(uuid.UUID)
	balance, err := c.withdrawalsService.GetUserBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := pkg.WriteJSON(w, http.StatusOK, balance); err != nil {
		c.logger.Error().Err(err).Msg("failed to write response")
	}
}

// CreateWithdrawal создаёт новое списание со счёта вознаграждений пользователя.
func (c *WithdrawalsController) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	var req schemas.CreateWithdrawalRequest
	if err := pkg.DecodeAndValidate(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(internal.UserKey{}).(uuid.UUID)
	withdrawal, err := c.withdrawalsService.CreateWithdrawal(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		// недостаточно средств
		if errors.Is(err, services.ErrInsufficientFunds) {
			http.Error(w, "insufficient funds", http.StatusPaymentRequired)
			return
		}
		// номер заказа некорректен
		if errors.Is(err, services.ErrInvalidOrderNumber) {
			http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := pkg.WriteJSON(w, http.StatusOK, withdrawal); err != nil {
		c.logger.Error().Err(err).Msg("failed to write response")
	}
}

// GetUserWithdrawals возвращает все списания пользователя.
func (c *WithdrawalsController) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(internal.UserKey{}).(uuid.UUID)
	withdrawals, err := c.withdrawalsService.GetUserWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := pkg.WriteJSON(w, http.StatusOK, withdrawals); err != nil {
		c.logger.Error().Err(err).Msg("failed to write response")
	}
}
