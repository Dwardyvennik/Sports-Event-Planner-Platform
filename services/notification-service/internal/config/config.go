package config

import sharedconfig "github.com/university/sports-event-planner-platform/pkg/config"

type Config = sharedconfig.Config

func Load() (Config, error) {
	return sharedconfig.Load(
		"notification-service",
		sharedconfig.WithGRPCAddr(":50053"),
		sharedconfig.WithHTTPAddr(":8083"),
		sharedconfig.WithPostgres("postgres://notification_user:notification_pass@localhost:5435/notification_db?sslmode=disable"),
		sharedconfig.WithRabbitMQ(),
	)
}
