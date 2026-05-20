package config

import (
	"os"

	sharedconfig "github.com/dwardyvennik/sports-event-planner-platform/pkg/config"
)

type MailgunConfig struct {
	APIKey string
	Domain string
	From   string
}

type NotificationConfig struct {
	sharedconfig.Config
	Mailgun MailgunConfig
}

func Load() (NotificationConfig, error) {
	base, err := sharedconfig.Load(
		"notification-service",
		sharedconfig.WithGRPCAddr(":50053"),
		sharedconfig.WithHTTPAddr(":8083"),
		sharedconfig.WithPostgres("postgres://notification_user:notification_pass@localhost:5435/notification_db?sslmode=disable"),
		sharedconfig.WithNATS(),
	)
	if err != nil {
		return NotificationConfig{}, err
	}

	return NotificationConfig{
		Config: base,
		Mailgun: MailgunConfig{
			APIKey: envString("MAILGUN_API_KEY", ""),
			Domain: envString("MAILGUN_DOMAIN", ""),
			From:   envString("MAILGUN_FROM", "noreply@example.com"),
		},
	}, nil
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
