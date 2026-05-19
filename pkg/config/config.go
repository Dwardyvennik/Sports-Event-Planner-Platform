package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App       AppConfig
	GRPC      GRPCConfig
	HTTP      HTTPConfig
	Postgres  PostgresConfig
	Redis     RedisConfig
	NATS      NATSConfig
	Endpoints ServiceEndpoints
}

type AppConfig struct {
	Name            string
	Environment     string
	LogLevel        string
	ShutdownTimeout time.Duration
}

type GRPCConfig struct {
	Addr string
}

type HTTPConfig struct {
	Addr string
}

type PostgresConfig struct {
	Enabled  bool
	URL      string
	MaxConns int32
	MinConns int32
}

type RedisConfig struct {
	Enabled  bool
	Addr     string
	Password string
	DB       int
}

type NATSConfig struct {
	Enabled bool
	URL     string
}

type ServiceEndpoints map[string]string

type Option func(*Config)

func Load(serviceName string, options ...Option) (Config, error) {
	cfg := Config{
		App: AppConfig{
			Name:            serviceName,
			Environment:     "development",
			LogLevel:        "info",
			ShutdownTimeout: 10 * time.Second,
		},
		GRPC: GRPCConfig{Addr: ":50051"},
		HTTP: HTTPConfig{Addr: ":8080"},
		Postgres: PostgresConfig{
			MaxConns: 10,
			MinConns: 1,
		},
		Redis: RedisConfig{
			Addr: "localhost:6379",
			DB:   0,
		},
		NATS: NATSConfig{
			URL: "nats://localhost:4222",
		},
		Endpoints: ServiceEndpoints{},
	}

	for _, option := range options {
		option(&cfg)
	}

	cfg.App.Name = envString("SERVICE_NAME", cfg.App.Name)
	cfg.App.Environment = envString("APP_ENV", cfg.App.Environment)
	cfg.App.LogLevel = envString("LOG_LEVEL", cfg.App.LogLevel)

	timeout, err := envDuration("SHUTDOWN_TIMEOUT", cfg.App.ShutdownTimeout)
	if err != nil {
		return cfg, err
	}
	cfg.App.ShutdownTimeout = timeout

	cfg.GRPC.Addr = envString("GRPC_ADDR", cfg.GRPC.Addr)
	cfg.HTTP.Addr = envString("HTTP_ADDR", cfg.HTTP.Addr)

	cfg.Postgres.URL = envString("DATABASE_URL", cfg.Postgres.URL)
	maxConns, err := envInt32("POSTGRES_MAX_CONNS", cfg.Postgres.MaxConns)
	if err != nil {
		return cfg, err
	}
	cfg.Postgres.MaxConns = maxConns

	minConns, err := envInt32("POSTGRES_MIN_CONNS", cfg.Postgres.MinConns)
	if err != nil {
		return cfg, err
	}
	cfg.Postgres.MinConns = minConns

	cfg.Redis.Addr = envString("REDIS_ADDR", cfg.Redis.Addr)
	cfg.Redis.Password = envString("REDIS_PASSWORD", cfg.Redis.Password)
	redisDB, err := envInt("REDIS_DB", cfg.Redis.DB)
	if err != nil {
		return cfg, err
	}
	cfg.Redis.DB = redisDB

	cfg.NATS.URL = envString("NATS_URL", cfg.NATS.URL)

	for name, addr := range cfg.Endpoints {
		cfg.Endpoints[name] = envString(endpointEnvName(name), addr)
	}

	if cfg.Postgres.Enabled && cfg.Postgres.URL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required when postgres is enabled")
	}
	if cfg.Redis.Enabled && cfg.Redis.Addr == "" {
		return cfg, fmt.Errorf("REDIS_ADDR is required when redis is enabled")
	}
	if cfg.NATS.Enabled && cfg.NATS.URL == "" {
		return cfg, fmt.Errorf("NATS_URL is required when nats is enabled")
	}

	return cfg, nil
}

func WithGRPCAddr(addr string) Option {
	return func(cfg *Config) {
		cfg.GRPC.Addr = addr
	}
}

func WithHTTPAddr(addr string) Option {
	return func(cfg *Config) {
		cfg.HTTP.Addr = addr
	}
}

func WithPostgres(defaultURL string) Option {
	return func(cfg *Config) {
		cfg.Postgres.Enabled = true
		cfg.Postgres.URL = defaultURL
	}
}

func WithRedis() Option {
	return func(cfg *Config) {
		cfg.Redis.Enabled = true
	}
}

func WithNATS() Option {
	return func(cfg *Config) {
		cfg.NATS.Enabled = true
	}
}

func WithEndpoint(name, addr string) Option {
	return func(cfg *Config) {
		cfg.Endpoints[name] = addr
	}
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

func envInt(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}

func envInt32(key string, fallback int32) (int32, error) {
	value, err := envInt(key, int(fallback))
	if err != nil {
		return 0, err
	}
	return int32(value), nil
}

func endpointEnvName(name string) string {
	switch name {
	case "auth":
		return "AUTH_SERVICE_GRPC_ADDR"
	case "event":
		return "EVENT_SERVICE_GRPC_ADDR"
	case "notification":
		return "NOTIFICATION_SERVICE_GRPC_ADDR"
	default:
		return "SERVICE_" + name + "_ADDR"
	}
}
