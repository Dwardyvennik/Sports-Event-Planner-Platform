package service

import (
	"context"
	"log/slog"

	"github.com/university/sports-event-planner-platform/pkg/health"
	"github.com/university/sports-event-planner-platform/services/event-service/internal/config"
	eventmemory "github.com/university/sports-event-planner-platform/services/event-service/internal/repository/memory"
	"github.com/university/sports-event-planner-platform/services/event-service/internal/usecase"
)

type Container struct {
	EventRepository *eventmemory.EventRepository
	EventUseCase    *usecase.EventUseCase
}

func NewContainer(ctx context.Context, cfg config.Config, log *slog.Logger) (*Container, error) {
	_ = ctx
	_ = cfg
	events := eventmemory.NewEventRepository()
	log.Info("event dependencies wired")

	return &Container{
		EventRepository: events,
		EventUseCase:    usecase.NewEventUseCase(events),
	}, nil
}

func (c *Container) Checks() map[string]health.Checker {
	return map[string]health.Checker{
		"usecase": c.EventUseCase.Health,
	}
}

func (c *Container) Close() error {
	return nil
}
