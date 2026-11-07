package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Polikarpov00/test_task_wallet/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {

	poolSettings, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	poolSettings.MaxConns = cfg.MaxConns
	poolSettings.MinConns = cfg.MinConns
	poolSettings.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolSettings)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = pool.Ping(pingCtx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pool: %w", err)
	}

	return pool, nil
}
