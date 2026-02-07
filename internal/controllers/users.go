package controllers

import (
	"errors"
	"net/http"

	"github.com/Pelfox/go-loyalty-system/internal"
	"github.com/Pelfox/go-loyalty-system/internal/schemas"
	"github.com/Pelfox/go-loyalty-system/internal/services"
	"github.com/Pelfox/go-loyalty-system/pkg"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// UsersController описывает обработчик (контроллер) для всех операций,
// связанных с существом пользователя.
type UsersController struct {
	logger       zerolog.Logger
	usersService *services.UsersService
}

// NewUsersController создаёт и возвращает новый UsersController.
func NewUsersController(
	logger zerolog.Logger,
	usersService *services.UsersService,
) *UsersController {
	return &UsersController{
		logger:       logger.With().Str("controller", "users").Logger(),
		usersService: usersService,
	}
}

// ApplyRoutes добавляет все пути данного обработчика к указанному роутеру.
func (c *UsersController) ApplyRoutes(router chi.Router) {
	router.Post("/register", c.Register)
	router.Post("/login", c.Login)
}

// setAuthenticationCookie устанавливает cookie аутентификации.
func (c *UsersController) setAuthenticationCookie(
	w http.ResponseWriter,
	token string,
	secure bool,
) {
	http.SetCookie(w, &http.Cookie{
		Name:     internal.SessionCookieName,
		Value:    token,
		Quoted:   false,
		Path:     "/",
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// Register регистрирует нового пользователя, используя данные из запроса.
func (c *UsersController) Register(w http.ResponseWriter, r *http.Request) {
	// декодируем полученный запрос
	var req schemas.RegisterUserRequest
	if err := pkg.DecodeAndValidate(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// создаём новый аккаунт пользователю, а так же токен аутентификации
	token, resp, err := c.usersService.Create(r.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrUserAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	c.setAuthenticationCookie(w, token, r.TLS != nil)
	if err := pkg.WriteJSON(w, http.StatusOK, resp); err != nil {
		c.logger.Error().Err(err).Msg("failed to write response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// Login проверяет правильность введённого пользователем пароля и авторизовывает его.
func (c *UsersController) Login(w http.ResponseWriter, r *http.Request) {
	// декодируем полученный запрос
	var request schemas.LoginUserRequest
	if err := pkg.DecodeAndValidate(r, &request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// проверяем корректность пароля пользователя и создаём токен аутентификации
	token, resp, err := c.usersService.Login(r.Context(), request)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) || errors.Is(err, services.ErrInvalidPassword) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	c.setAuthenticationCookie(w, token, r.TLS != nil)
	if err := pkg.WriteJSON(w, http.StatusOK, resp); err != nil {
		c.logger.Error().Err(err).Msg("failed to write response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
