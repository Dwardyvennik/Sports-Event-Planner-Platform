package natsx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/university/sports-event-planner-platform/pkg/config"
	"github.com/university/sports-event-planner-platform/pkg/health"
)

func Connect(ctx context.Context, cfg config.NATSConfig) (*nats.Conn, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	var conn *nats.Conn
	var err error
	done := make(chan struct{})
	go func() {
		conn, err = nats.Connect(
			cfg.URL,
			nats.Name("sports-event-planner"),
			nats.Timeout(5*time.Second),
			nats.RetryOnFailedConnect(true),
			nats.MaxReconnects(-1),
			nats.ReconnectWait(2*time.Second),
		)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		if err != nil {
			return nil, fmt.Errorf("connect nats: %w", err)
		}
		return conn, nil
	}
}

func HealthCheck(conn *nats.Conn) health.Checker {
	return func(context.Context) error {
		if conn == nil {
			return nil
		}
		if conn.IsClosed() {
			return errors.New("nats connection is closed")
		}
		if !conn.IsConnected() && !conn.IsReconnecting() {
			return errors.New("nats connection is not connected")
		}
		return nil
	}
}

func Drain(conn *nats.Conn) error {
	if conn == nil {
		return nil
	}
	return conn.Drain()
}
