package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфигурации БД: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 2
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctxWithTimeout, config)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания пула соединений БД: %w", err)
	}

	if err := pool.Ping(ctxWithTimeout); err != nil {
		return nil, fmt.Errorf("база данных недоступна при пинге: %w", err)
	}

	return pool, nil
}
