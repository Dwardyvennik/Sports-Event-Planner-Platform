package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/university/sports-event-planner-platform/pkg/config"
	"github.com/university/sports-event-planner-platform/pkg/health"
)

func Connect(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

func HealthCheck(pool *pgxpool.Pool) health.Checker {
	return func(ctx context.Context) error {
		if pool == nil {
			return nil
		}
		return pool.Ping(ctx)
	}
}
