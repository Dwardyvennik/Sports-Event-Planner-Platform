package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"github.com/university/sports-event-planner-platform/pkg/health"
	"github.com/university/sports-event-planner-platform/pkg/natsx"
	sharedpostgres "github.com/university/sports-event-planner-platform/pkg/postgres"
	"github.com/university/sports-event-planner-platform/services/notification-service/internal/config"
	"github.com/university/sports-event-planner-platform/services/notification-service/internal/consumer"
	"github.com/university/sports-event-planner-platform/services/notification-service/internal/mailgun"
	notificationpostgres "github.com/university/sports-event-planner-platform/services/notification-service/internal/repository/postgres"
	"github.com/university/sports-event-planner-platform/services/notification-service/internal/usecase"
)

type Container struct {
	DB                     *pgxpool.Pool
	NATS                   *nats.Conn
	NotificationRepository *notificationpostgres.NotificationRepository
	NotificationUseCase    *usecase.NotificationUseCase
	EventConsumer          *consumer.EventConsumer
}

func NewContainer(ctx context.Context, cfg config.NotificationConfig, log *slog.Logger) (*Container, error) {
	db, err := sharedpostgres.Connect(ctx, cfg.Postgres)
	if err != nil {
		return nil, err
	}

	broker, err := natsx.Connect(ctx, cfg.NATS)
	if err != nil {
		log.Warn("nats unavailable, event consumer disabled", "error", err)
		broker = nil
	}

	mg := mailgun.NewClient(cfg.Mailgun.APIKey, cfg.Mailgun.Domain, cfg.Mailgun.From)

	notifications := notificationpostgres.NewNotificationRepository(db)
	uc := usecase.NewNotificationUseCase(notifications, mg, log)

	ec := consumer.NewEventConsumer(broker, uc, log)

	log.Info("notification dependencies wired")

	return &Container{
		DB:                     db,
		NATS:                   broker,
		NotificationRepository: notifications,
		NotificationUseCase:    uc,
		EventConsumer:          ec,
	}, nil
}

func (c *Container) Checks() map[string]health.Checker {
	return map[string]health.Checker{
		"postgres": sharedpostgres.HealthCheck(c.DB),
		"nats":     natsx.HealthCheck(c.NATS),
		"usecase":  c.NotificationUseCase.Health,
	}
}

func (c *Container) Close() error {
	var err error
	if c.NATS != nil {
		err = errors.Join(err, natsx.Drain(c.NATS))
	}
	if c.DB != nil {
		c.DB.Close()
	}
	return err
}
