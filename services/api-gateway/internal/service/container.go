package service

import (
	"context"
	"log/slog"

	"github.com/dwardyvennik/sports-event-planner-platform/pkg/health"
	"github.com/dwardyvennik/sports-event-planner-platform/services/api-gateway/internal/config"
	deliverygrpc "github.com/dwardyvennik/sports-event-planner-platform/services/api-gateway/internal/delivery/grpc"
	"github.com/dwardyvennik/sports-event-planner-platform/services/api-gateway/internal/domain"
	"github.com/dwardyvennik/sports-event-planner-platform/services/api-gateway/internal/usecase"
)

type Container struct {
	Clients        *deliverygrpc.Clients
	GatewayUseCase *usecase.GatewayUseCase
}

func NewContainer(ctx context.Context, cfg config.Config, log *slog.Logger) (*Container, error) {
	clients, err := deliverygrpc.Dial(ctx, cfg.Endpoints, log)
	if err != nil {
		return nil, err
	}

	upstreams := make([]domain.Upstream, 0, len(cfg.Endpoints))
	for name, address := range cfg.Endpoints {
		upstreams = append(upstreams, domain.Upstream{Name: name, Address: address})
	}

	log.Info("gateway dependencies wired")
	return &Container{
		Clients:        clients,
		GatewayUseCase: usecase.NewGatewayUseCase(upstreams),
	}, nil
}

func (c *Container) Checks() map[string]health.Checker {
	return c.Clients.Checks()
}

func (c *Container) Close() error {
	return c.Clients.Close()
}
