package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/university/sports-event-planner-platform/pkg/httpx"
	"github.com/university/sports-event-planner-platform/pkg/lifecycle"
	"github.com/university/sports-event-planner-platform/pkg/logger"
	serviceconfig "github.com/university/sports-event-planner-platform/services/api-gateway/internal/config"
	deliveryhttp "github.com/university/sports-event-planner-platform/services/api-gateway/internal/delivery/http"
	"github.com/university/sports-event-planner-platform/services/api-gateway/internal/service"
)

func main() {
	cfg, err := serviceconfig.Load()
	if err != nil {
		slog.Default().Error("load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.App)
	ctx, stop := lifecycle.SignalContext()
	defer stop()

	container, err := service.NewContainer(ctx, cfg, log)
	if err != nil {
		log.Error("wire dependencies", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := container.Close(); err != nil {
			log.Error("close dependencies", "error", err)
		}
	}()

	httpServer := httpx.NewServer(cfg.HTTP, cfg.App.Name, container.Checks(), log, func(mux *http.ServeMux) {
		deliveryhttp.RegisterRoutes(mux, container.GatewayUseCase)
	})

	errCh := make(chan error, 1)
	go func() { errCh <- httpx.Serve(httpServer, log) }()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		log.Error("server stopped unexpectedly", "error", err)
		stop()
	}

	shutdownCtx, cancel := lifecycle.ShutdownContext(cfg.App.ShutdownTimeout)
	defer cancel()

	if err := httpx.Shutdown(shutdownCtx, httpServer, log); err != nil {
		log.Error("http shutdown failed", "error", err)
	}
}
