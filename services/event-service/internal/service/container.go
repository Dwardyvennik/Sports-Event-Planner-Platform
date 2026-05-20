package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"github.com/dwardyvennik/sports-event-planner-platform/pkg/health"
	"github.com/dwardyvennik/sports-event-planner-platform/pkg/natsx"
	sharedpostgres "github.com/dwardyvennik/sports-event-planner-platform/pkg/postgres"
	"github.com/dwardyvennik/sports-event-planner-platform/pkg/redisx"
	"github.com/dwardyvennik/sports-event-planner-platform/services/event-service/internal/config"
	eventpostgres "github.com/dwardyvennik/sports-event-planner-platform/services/event-service/internal/repository/postgres"
	"github.com/dwardyvennik/sports-event-planner-platform/services/event-service/internal/usecase"
)

type Container struct {
	DB              *pgxpool.Pool
	Redis           *redis.Client
	NATS            *nats.Conn
	EventRepository *eventpostgres.EventRepository
	EventUseCase    *usecase.EventUseCase
}

func NewContainer(ctx context.Context, cfg config.Config, log *slog.Logger) (*Container, error) {
	db, err := sharedpostgres.Connect(ctx, cfg.Postgres)
	if err != nil {
		return nil, err
	}

	cache, err := redisx.Connect(ctx, cfg.Redis)
	if err != nil {
		log.Warn("redis unavailable, event cache disabled", "error", err)
		cache = nil
	}

	broker, err := natsx.Connect(ctx, cfg.NATS)
	if err != nil {
		log.Warn("nats unavailable, event publisher disabled", "error", err)
		broker = nil
	}

	events := eventpostgres.NewEventRepository(db)
	log.Info("event dependencies wired")

	return &Container{
		DB:              db,
		Redis:           cache,
		NATS:            broker,
		EventRepository: events,
		EventUseCase:    usecase.NewEventUseCase(events, cache, broker, log),
	}, nil
}

func (c *Container) Checks() map[string]health.Checker {
	return map[string]health.Checker{
		"postgres": sharedpostgres.HealthCheck(c.DB),
		"redis":    redisx.HealthCheck(c.Redis),
		"nats":     natsx.HealthCheck(c.NATS),
		"usecase":  c.EventUseCase.Health,
	}
}

func (c *Container) Close() error {
	var err error
	if c.Redis != nil {
		err = errors.Join(err, c.Redis.Close())
	}
	if c.NATS != nil {
		err = errors.Join(err, natsx.Drain(c.NATS))
	}
	if c.DB != nil {
		c.DB.Close()
	}
	return err
}
