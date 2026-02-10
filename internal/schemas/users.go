package schemas

import "github.com/Pelfox/go-loyalty-system/internal/models"

// RegisterUserRequest описывает возможный запрос на регистрацию пользователя
type RegisterUserRequest struct {
	// Login - уникальный логин пользователя.
	Login string `json:"login" validate:"required"`
	// Password - plaintext пароль пользователя для хэширования.
	Password string `json:"password" validate:"required,min=8,max=64"`
}

// RegisterUserResponse описывает ответ на регистрацию пользователя.
type RegisterUserResponse struct {
	*models.User
}

// LoginUserRequest описывает содержимое запроса на вход в аккаунт.
type LoginUserRequest struct {
	// Login - уникальный логин пользователя.
	Login string `json:"login" validate:"required"`
	// Password - пароль пользователя.
	Password string `json:"password" validate:"required,min=8,max=64"`
}

// LoginUserResponse описывает ответ на запрос входа в аккаунт.
type LoginUserResponse struct {
	*models.User
}
