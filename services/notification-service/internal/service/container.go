package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/university/sports-event-planner-platform/pkg/health"
	sharedpostgres "github.com/university/sports-event-planner-platform/pkg/postgres"
	"github.com/university/sports-event-planner-platform/pkg/rabbitmq"
	"github.com/university/sports-event-planner-platform/services/notification-service/internal/config"
	notificationpostgres "github.com/university/sports-event-planner-platform/services/notification-service/internal/repository/postgres"
	"github.com/university/sports-event-planner-platform/services/notification-service/internal/usecase"
)

type Container struct {
	DB                     *pgxpool.Pool
	RabbitMQ               *amqp.Connection
	NotificationRepository *notificationpostgres.NotificationRepository
	NotificationUseCase    *usecase.NotificationUseCase
}

func NewContainer(ctx context.Context, cfg config.Config, log *slog.Logger) (*Container, error) {
	db, err := sharedpostgres.Connect(ctx, cfg.Postgres)
	if err != nil {
		return nil, err
	}

	broker, err := rabbitmq.Connect(ctx, cfg.RabbitMQ)
	if err != nil {
		if db != nil {
			db.Close()
		}
		return nil, err
	}

	notifications := notificationpostgres.NewNotificationRepository(db)
	log.Info("notification dependencies wired")

	return &Container{
		DB:                     db,
		RabbitMQ:               broker,
		NotificationRepository: notifications,
		NotificationUseCase:    usecase.NewNotificationUseCase(notifications),
	}, nil
}

func (c *Container) Checks() map[string]health.Checker {
	return map[string]health.Checker{
		"postgres": sharedpostgres.HealthCheck(c.DB),
		"rabbitmq": rabbitmq.HealthCheck(c.RabbitMQ),
		"usecase":  c.NotificationUseCase.Health,
	}
}

func (c *Container) Close() error {
	var err error
	if c.RabbitMQ != nil {
		err = errors.Join(err, c.RabbitMQ.Close())
	}
	if c.DB != nil {
		c.DB.Close()
	}
	return err
}
