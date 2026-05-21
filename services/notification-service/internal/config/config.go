package config

import (
	"fmt"
	"os"
	"strconv"

	sharedconfig "github.com/dwardyvennik/sports-event-planner-platform/pkg/config"
)

type MailgunConfig struct {
	APIKey string
	Domain string
	From   string
}

type SMTPConfig struct {
	Host string
	Port string
	From string
	TLS  bool
}

type NotificationConfig struct {
	sharedconfig.Config
	Mailgun                  MailgunConfig
	SMTP                     SMTPConfig
	EmailChannelEnabled      bool
	EventNotificationChannel string
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

	smtpTLS, err := envBool("SMTP_TLS", false)
	if err != nil {
		return NotificationConfig{}, err
	}
	emailChannelEnabled, err := envBool("EMAIL_CHANNEL_ENABLED", false)
	if err != nil {
		return NotificationConfig{}, err
	}

	return NotificationConfig{
		Config: base,
		Mailgun: MailgunConfig{
			APIKey: envString("MAILGUN_API_KEY", ""),
			Domain: envString("MAILGUN_DOMAIN", ""),
			From:   envString("MAILGUN_FROM", "noreply@sports-platform.local"),
		},
		SMTP: SMTPConfig{
			Host: envString("SMTP_HOST", ""),
			Port: envString("SMTP_PORT", "1025"),
			From: envString("SMTP_FROM", "noreply@sports-platform.local"),
			TLS:  smtpTLS,
		},
		EmailChannelEnabled:      emailChannelEnabled,
		EventNotificationChannel: envString("EVENT_NOTIFICATION_CHANNEL", "mock"),
	}, nil
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) (bool, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}
