package config

import sharedconfig "github.com/university/sports-event-planner-platform/pkg/config"

type Config = sharedconfig.Config

func Load() (Config, error) {
	return sharedconfig.Load(
		"auth-service",
		sharedconfig.WithGRPCAddr(":50051"),
		sharedconfig.WithHTTPAddr(":8081"),
		sharedconfig.WithPostgres("postgres://auth_user:auth_pass@localhost:5433/auth_db?sslmode=disable"),
		sharedconfig.WithRedis(),
	)
}
