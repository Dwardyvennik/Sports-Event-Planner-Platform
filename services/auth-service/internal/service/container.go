package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/dwardyvennik/sports-event-planner-platform/pkg/health"
	sharedpostgres "github.com/dwardyvennik/sports-event-planner-platform/pkg/postgres"
	"github.com/dwardyvennik/sports-event-planner-platform/pkg/redisx"
	"github.com/dwardyvennik/sports-event-planner-platform/services/auth-service/internal/config"
	authpostgres "github.com/dwardyvennik/sports-event-planner-platform/services/auth-service/internal/repository/postgres"
	"github.com/dwardyvennik/sports-event-planner-platform/services/auth-service/internal/usecase"
)

type Container struct {
	DB             *pgxpool.Pool
	Redis          *redis.Client
	UserRepository *authpostgres.UserRepository
	AuthUseCase    *usecase.AuthUseCase
}

func NewContainer(ctx context.Context, cfg config.Config, log *slog.Logger) (*Container, error) {
	db, err := sharedpostgres.Connect(ctx, cfg.Postgres)
	if err != nil {
		return nil, err
	}

	cache, err := redisx.Connect(ctx, cfg.Redis)
	if err != nil {
		log.Warn("redis unavailable, auth profile cache disabled", "error", err)
		cache = nil
	}

	users := authpostgres.NewUserRepository(db)
	log.Info("auth dependencies wired")

	return &Container{
		DB:             db,
		Redis:          cache,
		UserRepository: users,
		AuthUseCase: usecase.NewAuthUseCase(users, cache, usecase.Config{
			JWTSecret:       cfg.Auth.JWTSecret,
			AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
			RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
		}),
	}, nil
}

func (c *Container) Checks() map[string]health.Checker {
	return map[string]health.Checker{
		"postgres": sharedpostgres.HealthCheck(c.DB),
		"redis":    redisx.HealthCheck(c.Redis),
		"usecase":  c.AuthUseCase.Health,
	}
}

func (c *Container) Close() error {
	var err error
	if c.Redis != nil {
		err = errors.Join(err, c.Redis.Close())
	}
	if c.DB != nil {
		c.DB.Close()
	}
	return err
}
