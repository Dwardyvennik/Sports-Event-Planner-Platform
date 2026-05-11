package config

import sharedconfig "github.com/university/sports-event-planner-platform/pkg/config"

type Config = sharedconfig.Config

func Load() (Config, error) {
	return sharedconfig.Load(
		"api-gateway",
		sharedconfig.WithHTTPAddr(":8080"),
		sharedconfig.WithEndpoint("auth", "localhost:50051"),
		sharedconfig.WithEndpoint("event", "localhost:50052"),
		sharedconfig.WithEndpoint("notification", "localhost:50053"),
	)
}
