package storage

import (
	"context"
	"time"

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
