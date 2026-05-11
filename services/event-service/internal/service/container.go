package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/university/sports-event-planner-platform/pkg/health"
	sharedpostgres "github.com/university/sports-event-planner-platform/pkg/postgres"
	"github.com/university/sports-event-planner-platform/pkg/rabbitmq"
	"github.com/university/sports-event-planner-platform/pkg/redisx"
	"github.com/university/sports-event-planner-platform/services/event-service/internal/config"
	eventpostgres "github.com/university/sports-event-planner-platform/services/event-service/internal/repository/postgres"
	"github.com/university/sports-event-planner-platform/services/event-service/internal/usecase"
)

type Container struct {
	DB              *pgxpool.Pool
	Redis           *redis.Client
	RabbitMQ        *amqp.Connection
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
		if db != nil {
			db.Close()
		}
		return nil, err
	}

	broker, err := rabbitmq.Connect(ctx, cfg.RabbitMQ)
	if err != nil {
		if cache != nil {
			_ = cache.Close()
		}
		if db != nil {
			db.Close()
		}
		return nil, err
	}

	events := eventpostgres.NewEventRepository(db)
	log.Info("event dependencies wired")

	return &Container{
		DB:              db,
		Redis:           cache,
		RabbitMQ:        broker,
		EventRepository: events,
		EventUseCase:    usecase.NewEventUseCase(events),
	}, nil
}

func (c *Container) Checks() map[string]health.Checker {
	return map[string]health.Checker{
		"postgres": sharedpostgres.HealthCheck(c.DB),
		"redis":    redisx.HealthCheck(c.Redis),
		"rabbitmq": rabbitmq.HealthCheck(c.RabbitMQ),
		"usecase":  c.EventUseCase.Health,
	}
}

func (c *Container) Close() error {
	var err error
	if c.RabbitMQ != nil {
		err = errors.Join(err, c.RabbitMQ.Close())
	}
	if c.Redis != nil {
		err = errors.Join(err, c.Redis.Close())
	}
	if c.DB != nil {
		c.DB.Close()
	}
	return err
}
