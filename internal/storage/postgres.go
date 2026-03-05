package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresConnSettings описывает возможные настройки подключения пула к
// серверу PostgreSQL.
type PostgresConnSettings struct {
	// MinConns - минимальное количество подключений пула.
	MinConns int32
	// MaxConns - максимально возможное количество подключений пула.
	MaxConns int32
	// MaxConnIdleTime - сколько соединение может простаивать.
	MaxConnIdleTime time.Duration
	// MaxConnLifetime - максимальное время "жизни" соединения.
	MaxConnLifetime time.Duration
}

// DefaultPostgresConnSettings несёт в себе стандартные значения подключения
// для пула PostgreSQL. Следует использовать эти значения по умолчанию.
var DefaultPostgresConnSettings = PostgresConnSettings{
	MinConns:        2,
	MaxConns:        10,
	MaxConnIdleTime: 30 * time.Minute,
	MaxConnLifetime: 1 * time.Hour,
}

// NewPostgresPool создаёт и возвращает пул подключений к серверу PostgreSQL по
// указанной строке подключения и настройкам пула.
func NewPostgresPool(
	ctx context.Context,
	connString string,
	settings PostgresConnSettings,
) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	// копируем значения из переданного settings
	cfg.MinConns = settings.MinConns
	cfg.MaxConns = settings.MaxConns
	cfg.MaxConnIdleTime = settings.MaxConnIdleTime
	cfg.MaxConnLifetime = settings.MaxConnLifetime

	return pgxpool.NewWithConfig(ctx, cfg)
}

// RunMigrations запускает и применяет все миграции из папки с миграциями.
func RunMigrations(connString string, migratePath string) error {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	defer driver.Close()

	migrationsDir := fmt.Sprintf("file://%s", filepath.ToSlash(migratePath))
	migrator, err := migrate.NewWithDatabaseInstance(migrationsDir, "postgres", driver)
	if err != nil {
		return err
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// isRetryable возвращает true, если переданная ошибка - одна из тех, что может
// быть повторена внутри одной транзакции.
func isRetryable(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "40001" || pgErr.Code == "40P01"
}

// MaxRetries - максимальное количество попыток повтора единой операции внутри
// одной транзакции.
const MaxRetries = 5

// txFn описывает принимаемый тип функции для WithTxRetry.
type txFn func(tx pgx.Tx) error

// WithTxRetry оборачивает переданную функцию в механизм повтора, в случае если
// происходит deadlock со стороны PostgreSQL. Функция возвращает последнюю
// полученную ошибку.
func WithTxRetry(ctx context.Context, pool *pgxpool.Pool, opts pgx.TxOptions, cb txFn) error {
	var lastErr error

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		// если контекст отменён - сразу же прекращаем попытки
		if err := ctx.Err(); err != nil {
			return err
		}

		tx, err := pool.BeginTx(ctx, opts)
		if err != nil {
			return err
		}

		// вызываем callback-функцию
		if err := cb(tx); err != nil {
			lastErr = err
			// пытаемся отменить транзакцию
			if err := tx.Rollback(ctx); err != nil {
				return err
			}

			// проверяем, стоит ли пытаться ещё раз повторить операцию
			if !isRetryable(err) {
				return err
			}

			continue
		}

		// пытаемся сделать коммит изменений
		if err := tx.Commit(ctx); err != nil {
			lastErr = err
			if !isRetryable(err) {
				return err
			}
			continue
		}

		return nil
	}

	return lastErr
}
