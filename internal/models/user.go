package models

import (
	"time"

	"github.com/google/uuid"
)

// User описывает сущность единого пользователя в базе данных.
type User struct {
	// ID - уникальный идентификатор для данного пользователя.
	ID uuid.UUID `json:"id"`
	// Login - уникальная строка, используемая пользователем для входа.
	Login string `json:"login"`
	// PasswordHash - криптографически хэшированный пароль.
	PasswordHash string `json:"-"`
	// CreatedAt отображает дату и время, когда пользователь был зарегистрирован.
	CreatedAt time.Time `json:"created_at"`
}
