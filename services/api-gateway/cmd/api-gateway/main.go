package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/dwardyvennik/sports-event-planner-platform/pkg/lifecycle"
	"github.com/dwardyvennik/sports-event-planner-platform/pkg/logger"
	serviceconfig "github.com/dwardyvennik/sports-event-planner-platform/services/api-gateway/internal/config"
	deliveryhttp "github.com/dwardyvennik/sports-event-planner-platform/services/api-gateway/internal/delivery/http"
	"github.com/dwardyvennik/sports-event-planner-platform/services/api-gateway/internal/service"
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

	router := deliveryhttp.NewRouter(container.Clients, container.Checks(), log)
	httpServer := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("http server listening", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		log.Error("server stopped unexpectedly", "error", err)
		stop()
	}

	shutdownCtx, cancel := lifecycle.ShutdownContext(cfg.App.ShutdownTimeout)
	defer cancel()

	log.Info("http server shutting down", "addr", httpServer.Addr)
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown failed", "error", err)
	}
}
