package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/dwardyvennik/sports-event-planner-platform/pkg/config"
	"github.com/dwardyvennik/sports-event-planner-platform/pkg/health"
)

func Connect(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

func HealthCheck(client *redis.Client) health.Checker {
	return func(ctx context.Context) error {
		if client == nil {
			return nil
		}
		return client.Ping(ctx).Err()
	}
}
