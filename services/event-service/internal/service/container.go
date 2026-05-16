package service

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/university/sports-event-planner-platform/pkg/health"
	sharedpostgres "github.com/university/sports-event-planner-platform/pkg/postgres"
	"github.com/university/sports-event-planner-platform/services/event-service/internal/config"
	eventpostgres "github.com/university/sports-event-planner-platform/services/event-service/internal/repository/postgres"
	"github.com/university/sports-event-planner-platform/services/event-service/internal/usecase"
)

type Container struct {
	DB              *pgxpool.Pool
	EventRepository *eventpostgres.EventRepository
	EventUseCase    *usecase.EventUseCase
}

func NewContainer(ctx context.Context, cfg config.Config, log *slog.Logger) (*Container, error) {
	db, err := sharedpostgres.Connect(ctx, cfg.Postgres)
	if err != nil {
		return nil, err
	}

	events := eventpostgres.NewEventRepository(db)
	log.Info("event dependencies wired")

	return &Container{
		DB:              db,
		EventRepository: events,
		EventUseCase:    usecase.NewEventUseCase(events),
	}, nil
}

func (c *Container) Checks() map[string]health.Checker {
	return map[string]health.Checker{
		"postgres": sharedpostgres.HealthCheck(c.DB),
		"usecase":  c.EventUseCase.Health,
	}
}

func (c *Container) Close() error {
	if c.DB != nil {
		c.DB.Close()
	}
	return nil
}
