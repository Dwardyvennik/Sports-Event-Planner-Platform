package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/university/sports-event-planner-platform/pkg/config"
	"github.com/university/sports-event-planner-platform/pkg/health"
)

func Connect(ctx context.Context, cfg config.RabbitMQConfig) (*amqp.Connection, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	dialer := &net.Dialer{}
	conn, err := amqp.DialConfig(cfg.URL, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
		Dial: func(network, address string) (net.Conn, error) {
			dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			return dialer.DialContext(dialCtx, network, address)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	return conn, nil
}

func HealthCheck(conn *amqp.Connection) health.Checker {
	return func(context.Context) error {
		if conn == nil {
			return nil
		}
		if conn.IsClosed() {
			return errors.New("rabbitmq connection is closed")
		}
		return nil
	}
}
