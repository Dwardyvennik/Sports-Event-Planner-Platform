package config

import sharedconfig "github.com/university/sports-event-planner-platform/pkg/config"

type Config = sharedconfig.Config

func Load() (Config, error) {
	return sharedconfig.Load(
		"event-service",
		sharedconfig.WithGRPCAddr(":50052"),
		sharedconfig.WithHTTPAddr(":8082"),
		sharedconfig.WithPostgres("postgres://event_user:event_pass@localhost:5434/event_db?sslmode=disable"),
	)
}
