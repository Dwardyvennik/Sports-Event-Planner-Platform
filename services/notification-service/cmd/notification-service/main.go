package main

import (
	"log/slog"
	"os"

	"github.com/dwardyvennik/sports-event-planner-platform/pkg/grpcx"
	"github.com/dwardyvennik/sports-event-planner-platform/pkg/httpx"
	"github.com/dwardyvennik/sports-event-planner-platform/pkg/lifecycle"
	"github.com/dwardyvennik/sports-event-planner-platform/pkg/logger"
	serviceconfig "github.com/dwardyvennik/sports-event-planner-platform/services/notification-service/internal/config"
	deliverygrpc "github.com/dwardyvennik/sports-event-planner-platform/services/notification-service/internal/delivery/grpc"
	"github.com/dwardyvennik/sports-event-planner-platform/services/notification-service/internal/service"
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

	go container.NotificationUseCase.ReminderWorker(ctx)
	go container.EventConsumer.Start(ctx)

	grpcServer := grpcx.NewServer(cfg.App.Name, log)
	deliverygrpc.Register(grpcServer, container.NotificationUseCase, log)

	httpServer := httpx.NewServer(cfg.HTTP, cfg.App.Name, container.Checks(), log, nil)
	errCh := make(chan error, 2)

	go func() { errCh <- grpcx.Serve(grpcServer, cfg.GRPC.Addr, log) }()
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
	grpcx.Shutdown(grpcServer, cfg.App.ShutdownTimeout, log)
}
