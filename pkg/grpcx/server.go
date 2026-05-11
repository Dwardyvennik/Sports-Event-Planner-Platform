package grpcx

import (
	"context"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewServer(service string, log *slog.Logger) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryLoggingInterceptor(log)),
	)

	healthServer := health.NewServer()
	healthServer.SetServingStatus(service, healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)
	reflection.Register(server)

	return server
}

func Serve(server *grpc.Server, addr string, log *slog.Logger) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Info("grpc server listening", "addr", addr)
	return server.Serve(listener)
}

func Shutdown(server *grpc.Server, timeout time.Duration, log *slog.Logger) {
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Info("grpc server stopped gracefully")
	case <-time.After(timeout):
		log.Warn("grpc graceful shutdown timed out; forcing stop")
		server.Stop()
	}
}

func unaryLoggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		started := time.Now()
		response, err := handler(ctx, req)
		if err != nil {
			log.WarnContext(ctx, "grpc request failed",
				"method", info.FullMethod,
				"duration_ms", time.Since(started).Milliseconds(),
				"error", err,
			)
			return response, err
		}
		log.DebugContext(ctx, "grpc request completed",
			"method", info.FullMethod,
			"duration_ms", time.Since(started).Milliseconds(),
		)
		return response, nil
	}
}
