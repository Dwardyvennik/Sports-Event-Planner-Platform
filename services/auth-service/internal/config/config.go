package config

import (
	"fmt"
	"os"
	"time"

	sharedconfig "github.com/dwardyvennik/sports-event-planner-platform/pkg/config"
)

type Config struct {
	sharedconfig.Config
	Auth AuthConfig
}

type AuthConfig struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load() (Config, error) {
	shared, err := sharedconfig.Load(
		"auth-service",
		sharedconfig.WithGRPCAddr(":50051"),
		sharedconfig.WithHTTPAddr(":8081"),
		sharedconfig.WithPostgres("postgres://auth_user:auth_pass@localhost:5433/auth_db?sslmode=disable"),
		sharedconfig.WithRedis(),
	)
	if err != nil {
		return Config{}, err
	}

	accessTTL, err := envDuration("JWT_ACCESS_TTL", time.Hour)
	if err != nil {
		return Config{}, err
	}
	refreshTTL, err := envDuration("JWT_REFRESH_TTL", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Config: shared,
		Auth: AuthConfig{
			JWTSecret:       envString("JWT_SECRET", "dev-auth-secret-change-me"),
			AccessTokenTTL:  accessTTL,
			RefreshTokenTTL: refreshTTL,
		},
	}.validate()
}

func (cfg Config) validate() (Config, error) {
	if cfg.Auth.JWTSecret == "" {
		return cfg, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.Auth.AccessTokenTTL <= 0 {
		return cfg, fmt.Errorf("JWT_ACCESS_TTL must be positive")
	}
	if cfg.Auth.RefreshTokenTTL <= 0 {
		return cfg, fmt.Errorf("JWT_REFRESH_TTL must be positive")
	}
	return cfg, nil
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}
