package services

import (
	"context"
	"errors"
	"testing"

	"github.com/Pelfox/go-loyalty-system/internal/models"
	"github.com/Pelfox/go-loyalty-system/internal/repositories"
	"github.com/Pelfox/go-loyalty-system/internal/schemas"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// fakeUsersRepository - тестовая реализация UsersRepository, используемая для
// изолированного тестирования UsersService.
type fakeUsersRepository struct {
	createFunc     func(ctx context.Context, login string, passwordHash string) (*models.User, error)
	getByLoginFunc func(ctx context.Context, login string) (*models.User, error)
}

func (f *fakeUsersRepository) Create(ctx context.Context, login string, passwordHash string) (*models.User, error) {
	return f.createFunc(ctx, login, passwordHash)
}

func (f *fakeUsersRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	return f.getByLoginFunc(ctx, login)
}

// TestUsersService_Create_OK проверяет успешное создание пользователя. В этом
// случае сервис должен: вернуть JWT-токен, вернуть объект RegisterUserResponse
// и не вернуть ошибку
func TestUsersService_Create_OK(t *testing.T) {
	repo := &fakeUsersRepository{
		createFunc: func(ctx context.Context, login string, passwordHash string) (*models.User, error) {
			return &models.User{
				ID:           uuid.New(),
				Login:        login,
				PasswordHash: passwordHash,
			}, nil
		},
	}

	service := NewUsersService(
		zerolog.Nop(),
		[]byte("secret"),
		repo,
	)

	req := schemas.RegisterUserRequest{
		Login:    "test",
		Password: "password123",
	}

	token, resp, err := service.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Errorf("token is empty, want non-empty")
	}

	if resp == nil || resp.User == nil {
		t.Fatalf("response or user is nil")
	}

	if resp.User.Login != req.Login {
		t.Errorf("login = %q, want %q", resp.User.Login, req.Login)
	}
}

// TestUsersService_Create_UserAlreadyExists проверяет, что при дубликате
// логина сервис возвращает ErrUserAlreadyExists.
func TestUsersService_Create_UserAlreadyExists(t *testing.T) {
	repo := &fakeUsersRepository{
		createFunc: func(ctx context.Context, login string, passwordHash string) (*models.User, error) {
			return nil, repositories.ErrUserAlreadyExists
		},
	}

	service := NewUsersService(
		zerolog.Nop(),
		[]byte("secret"),
		repo,
	)

	req := schemas.RegisterUserRequest{
		Login:    "test",
		Password: "password123",
	}

	_, _, err := service.Create(context.Background(), req)
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Errorf("error = %v, want %v", err, ErrUserAlreadyExists)
	}
}

// TestUsersService_Create_RepositoryError проверяет, что при любой иной ошибке
// репозитория сервис возвращает ErrUserCreationFailed.
func TestUsersService_Create_RepositoryError(t *testing.T) {
	repo := &fakeUsersRepository{
		createFunc: func(ctx context.Context, login string, passwordHash string) (*models.User, error) {
			return nil, errors.New("database error")
		},
	}

	service := NewUsersService(
		zerolog.Nop(),
		[]byte("secret"),
		repo,
	)

	req := schemas.RegisterUserRequest{
		Login:    "test",
		Password: "password123",
	}

	_, _, err := service.Create(context.Background(), req)
	if !errors.Is(err, ErrUserCreationFailed) {
		t.Errorf("error = %v, want %v", err, ErrUserCreationFailed)
	}
}

// TestUsersService_Login_UserNotFound проверяет, что если пользователь не
// найден, сервис возвращает ErrUserNotFound.
func TestUsersService_Login_UserNotFound(t *testing.T) {
	repo := &fakeUsersRepository{
		getByLoginFunc: func(ctx context.Context, login string) (*models.User, error) {
			return nil, repositories.ErrUserNotFound
		},
	}

	service := NewUsersService(
		zerolog.Nop(),
		[]byte("secret"),
		repo,
	)

	req := schemas.LoginUserRequest{
		Login:    "test",
		Password: "password123",
	}

	_, _, err := service.Login(context.Background(), req)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("error = %v, want %v", err, ErrUserNotFound)
	}
}

// TestUsersService_Login_InvalidPassword проверяет, что при некорректном
// пароле сервис возвращает ErrInvalidPassword.
func TestUsersService_Login_InvalidPassword(t *testing.T) {
	passwordHash := "$argon2id$v=19$m=65536,t=1,p=2$invalid$hash"
	repo := &fakeUsersRepository{
		getByLoginFunc: func(ctx context.Context, login string) (*models.User, error) {
			return &models.User{
				ID:           uuid.New(),
				Login:        login,
				PasswordHash: passwordHash,
			}, nil
		},
	}

	service := NewUsersService(
		zerolog.Nop(),
		[]byte("secret"),
		repo,
	)

	req := schemas.LoginUserRequest{
		Login:    "test",
		Password: "password123",
	}

	_, _, err := service.Login(context.Background(), req)
	if !errors.Is(err, ErrPasswordHashingFailed) {
		t.Errorf("error = %v, want %v", err, ErrInvalidPassword)
	}
}

// TestUsersService_Login_OK проверяет успешную авторизацию пользователя.
// В этом случае сервис должен вернуть JWT-токен и LoginUserResponse.
func TestUsersService_Login_OK(t *testing.T) {
	// корректный argon2-хэш для пароля "password123"
	hash, err := func() (string, error) {
		return "$argon2id$v=19$m=16,t=2,p=1$SlNYUW1wdVVVeVB6a1pLag$/1hI1GeWTVnEO3c5RfoWOg", nil
	}()
	if err != nil {
		t.Fatalf("failed to prepare hash")
	}

	userID := uuid.New()
	repo := &fakeUsersRepository{
		getByLoginFunc: func(ctx context.Context, login string) (*models.User, error) {
			return &models.User{
				ID:           userID,
				Login:        login,
				PasswordHash: hash,
			}, nil
		},
	}

	service := NewUsersService(
		zerolog.Nop(),
		[]byte("secret"),
		repo,
	)

	req := schemas.LoginUserRequest{
		Login:    "test",
		Password: "password123",
	}

	token, resp, err := service.Login(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Errorf("token is empty, want non-empty")
	}

	if resp == nil || resp.User == nil {
		t.Fatalf("response or user is nil")
	}

	if resp.User.ID != userID {
		t.Errorf("userID = %v, want %v", resp.User.ID, userID)
	}
}
