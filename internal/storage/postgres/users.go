package postgres

import (
	"context"
	"errors"

	"github.com/Pelfox/go-loyalty-system/internal/models"
	"github.com/Pelfox/go-loyalty-system/internal/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UsersRepository реализует Postgres-совместимый репозиторий для пользователей.
type UsersRepository struct {
	db *pgxpool.Pool
}

// NewUsersRepository создаёт и возвращает новый UsersRepository, с указанным
// пулом подключений PostgreSQL.
func NewUsersRepository(db *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) Create(ctx context.Context, login string, passwordHash string) (*models.User, error) {
	var user models.User
	user.Login = login
	user.PasswordHash = passwordHash

	query := "INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, created_at"
	err := r.db.QueryRow(ctx, query, login, passwordHash).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, repositories.ErrUserAlreadyExists
		}
		return nil, err
	}

	return &user, nil
}

func (r *UsersRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	query := "SELECT * FROM users WHERE login = $1"
	err := r.db.QueryRow(ctx, query, login).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
