package services

import (
	"context"
	"errors"
	"time"

	"github.com/Pelfox/go-loyalty-system/internal/repositories"
	"github.com/Pelfox/go-loyalty-system/internal/schemas"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// tokenLifetime описывает время "жизни" токена (пока он валиден).
const tokenLifetime = 1 * time.Hour

var (
	// ErrPasswordHashingFailed описывает ошибку, когда не удаётся захэшировать
	// пользовательский пароль.
	ErrPasswordHashingFailed = errors.New("failed to hash the password")
	// ErrUserAlreadyExists описывает ошибку, когда пользователь с данным
	// логином уже существует.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrUserCreationFailed описывает ошибку, при которой что-то пошло не так
	// при создании пользователя в базе данных.
	ErrUserCreationFailed = errors.New("failed to create user")
	// ErrUserNotFound используется, когда пользователь с указанным логином не
	// был найден.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserLookupFailed описывает ошибку, при которой не удаётся получить
	// информацию о пользователе.
	ErrUserLookupFailed = errors.New("failed to lookup user")
	// ErrInvalidPassword описывает ошибку, возникающую при некорректном вводе
	// пароля пользователем.
	ErrInvalidPassword = errors.New("invalid password")
	// ErrTokenCreationFailed описывает ошибку, при которой невозможно создать
	// JWT-токен для аутентификации.
	ErrTokenCreationFailed = errors.New("failed to create authentication token")
)

// UsersService реализовывает весь пользовательский функционал.
type UsersService struct {
	logger          zerolog.Logger
	jwtSecret       []byte
	usersRepository repositories.UsersRepository
}

// NewUsersService создаёт и возвращает новый UsersService.
func NewUsersService(
	logger zerolog.Logger,
	jwtSecret []byte,
	usersRepository repositories.UsersRepository,
) *UsersService {
	return &UsersService{
		logger:          logger.With().Str("service", "users").Logger(),
		jwtSecret:       jwtSecret,
		usersRepository: usersRepository,
	}
}

// createAuthenticationToken создаёт новый JWT-токен для аутентификации для
// пользователя с переданным ID. Токен истекает через tokenLifetime времени.
func (s *UsersService) createAuthenticationToken(userID uuid.UUID) (string, error) {
	currentTime := time.Now()
	claims := &jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(currentTime.Add(tokenLifetime)),
		IssuedAt:  jwt.NewNumericDate(currentTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// Create создаёт нового пользователя, исходя из указанного логина и пароля.
// Возвращает ошибку, в случае дубликата логина или иной ошибки в системе. В
// ином случае - JWT-токен для аутентификации и объект schemas.RegisterUserResponse.
func (s *UsersService) Create(
	ctx context.Context,
	req schemas.RegisterUserRequest,
) (string, *schemas.RegisterUserResponse, error) {
	// хэшируем пароль пользователя
	hashedPassword, err := argon2id.CreateHash(req.Password, argon2id.DefaultParams)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to hash the password")
		return "", nil, ErrPasswordHashingFailed
	}

	// создаём новый объект пользователя
	user, err := s.usersRepository.Create(ctx, req.Login, hashedPassword)
	if err != nil {
		if errors.Is(err, repositories.ErrUserAlreadyExists) {
			return "", nil, ErrUserAlreadyExists
		}
		s.logger.Error().Err(err).Msg("failed to create user")
		return "", nil, ErrUserCreationFailed
	}

	// создаём токен для аутентификации
	token, err := s.createAuthenticationToken(user.ID)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create auth token")
		return "", nil, ErrTokenCreationFailed
	}

	return token, &schemas.RegisterUserResponse{User: user}, nil
}

// Login проверяет корректность пароля, и авторизовывает пользователя. В
// случае, если пользователь не найден, пароль некорректен или иной другой
// системной ошибки, возвращает её. Иначе - JWT-токен для аутентификации и
// объект schemas.LoginUserResponse.
func (s *UsersService) Login(
	ctx context.Context,
	req schemas.LoginUserRequest,
) (string, *schemas.LoginUserResponse, error) {
	// ищем пользователя по указанному логину
	user, err := s.usersRepository.GetByLogin(ctx, req.Login)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return "", nil, ErrUserNotFound
		}
		s.logger.Error().Err(err).Msg("failed to get user")
		return "", nil, ErrUserLookupFailed
	}

	// подтверждаем валидность пароля пользователя
	isValid, err := argon2id.ComparePasswordAndHash(req.Password, user.PasswordHash)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to compare password")
		return "", nil, ErrPasswordHashingFailed
	}
	if !isValid {
		return "", nil, ErrInvalidPassword
	}

	// создаём токен для аутентификации
	token, err := s.createAuthenticationToken(user.ID)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to create auth token")
		return "", nil, ErrTokenCreationFailed
	}

	return token, &schemas.LoginUserResponse{User: user}, nil
}
