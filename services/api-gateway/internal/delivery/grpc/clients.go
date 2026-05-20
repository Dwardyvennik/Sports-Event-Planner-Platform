package grpc

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/dwardyvennik/sports-event-planner-platform/pkg/grpcx"
	"github.com/dwardyvennik/sports-event-planner-platform/pkg/health"
	authv1 "github.com/dwardyvennik/sports-event-planner-platform/services/auth-service/proto/auth/v1"
	eventv1 "github.com/dwardyvennik/sports-event-planner-platform/services/event-service/proto/event/v1"
	notificationv1 "github.com/dwardyvennik/sports-event-planner-platform/services/notification-service/proto/notification/v1"
)

type Clients struct {
	conns        map[string]*grpc.ClientConn
	Auth         authv1.AuthServiceClient
	Event        eventv1.EventServiceClient
	Notification notificationv1.NotificationServiceClient
}

func Dial(_ context.Context, endpoints map[string]string, log *slog.Logger) (*Clients, error) {
	conns := make(map[string]*grpc.ClientConn, len(endpoints))
	for name, addr := range endpoints {
		conn, err := grpc.NewClient(
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.CallContentSubtype(grpcx.JSONCodecName)),
			grpc.WithChainUnaryInterceptor(
				grpcx.UnaryClientTimeoutInterceptor(3*time.Second),
				grpcx.UnaryClientLoggingInterceptor(log),
			),
		)
		if err != nil {
			return nil, err
		}
		conns[name] = conn
	}

	return &Clients{
		conns:        conns,
		Auth:         authv1.NewAuthServiceClient(conns["auth"]),
		Event:        eventv1.NewEventServiceClient(conns["event"]),
		Notification: notificationv1.NewNotificationServiceClient(conns["notification"]),
	}, nil
}

func (c *Clients) Checks() map[string]health.Checker {
	checks := make(map[string]health.Checker, len(c.conns))
	for name, conn := range c.conns {
		conn := conn
		checks["grpc_"+name] = func(context.Context) error {
			if conn.GetState() == connectivity.Shutdown {
				return errors.New("grpc connection is shutdown")
			}
			return nil
		}
	}
	return checks
}

func (c *Clients) Close() error {
	var err error
	for _, conn := range c.conns {
		err = errors.Join(err, conn.Close())
	}
	return err
}
