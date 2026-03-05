package repositories

import (
	"context"

	"github.com/Pelfox/go-loyalty-system/internal/models"
)

// UsersRepository описывает возможный набор действий с моделями пользователей.
type UsersRepository interface {
	// Create создаёт нового пользователя, и возвращает его модель. Вернёт
	// ErrUserAlreadyExists, если пользователь с данным логином уже зарегистрирован.
	Create(ctx context.Context, login string, passwordHash string) (*models.User, error)
	// GetByLogin возвращает модель пользователя, используя указанный логин. В
	// случае, если пользователь найден не будет, вернёт ErrUserNotFound.
	GetByLogin(ctx context.Context, login string) (*models.User, error)
}
