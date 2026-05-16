package config

import sharedconfig "github.com/university/sports-event-planner-platform/pkg/config"

type Config = sharedconfig.Config

func Load() (Config, error) {
	return sharedconfig.Load(
		"event-service",
		sharedconfig.WithGRPCAddr(":50052"),
		sharedconfig.WithHTTPAddr(":8082"),
	)
}
