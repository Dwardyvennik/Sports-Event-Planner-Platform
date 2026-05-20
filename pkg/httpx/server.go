package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dwardyvennik/sports-event-planner-platform/pkg/config"
	"github.com/dwardyvennik/sports-event-planner-platform/pkg/health"
	"github.com/dwardyvennik/sports-event-planner-platform/pkg/metrics"
)

func NewServer(cfg config.HTTPConfig, service string, checks map[string]health.Checker, log *slog.Logger, routes func(*http.ServeMux)) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health.Liveness(service))
	mux.HandleFunc("/readyz", health.Readiness(service, checks, log))
	mux.Handle("/metrics", metrics.Handler(service))

	if routes != nil {
		routes(mux)
	}

	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           metrics.InstrumentHTTP(service, requestLogger(log, mux)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func Serve(server *http.Server, log *slog.Logger) error {
	log.Info("http server listening", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func Shutdown(ctx context.Context, server *http.Server, log *slog.Logger) error {
	log.Info("http server shutting down", "addr", server.Addr)
	return server.Shutdown(ctx)
}

func requestLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.DebugContext(r.Context(), "http request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}
