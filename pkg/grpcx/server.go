package grpcx

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	grpcServerRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Subsystem: "grpc_server",
		Name:      "requests_total",
		Help:      "Total number of gRPC server requests.",
	}, []string{"service", "method", "code"})

	grpcServerDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "sports_planner",
		Subsystem: "grpc_server",
		Name:      "request_duration_seconds",
		Help:      "gRPC server request duration.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"service", "method"})

	grpcClientRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Subsystem: "grpc_client",
		Name:      "requests_total",
		Help:      "Total number of gRPC client requests.",
	}, []string{"method", "code"})

	grpcClientDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "sports_planner",
		Subsystem: "grpc_client",
		Name:      "request_duration_seconds",
		Help:      "gRPC client request duration.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method"})
)

func NewServer(service string, log *slog.Logger) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryInterceptor(log),
			unaryMetricsInterceptor(service),
			unaryLoggingInterceptor(log),
		),
	)

	healthServer := health.NewServer()
	healthServer.SetServingStatus(service, healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)
	reflection.Register(server)

	return server
}

func UnaryClientTimeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if _, ok := ctx.Deadline(); !ok && timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func UnaryClientLoggingInterceptor(log *slog.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		started := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		code := status.Code(err).String()
		grpcClientRequests.WithLabelValues(method, code).Inc()
		grpcClientDuration.WithLabelValues(method).Observe(time.Since(started).Seconds())
		if err != nil {
			log.WarnContext(ctx, "grpc client request failed",
				"method", method,
				"duration_ms", time.Since(started).Milliseconds(),
				"error", err,
			)
			return err
		}
		log.DebugContext(ctx, "grpc client request completed",
			"method", method,
			"duration_ms", time.Since(started).Milliseconds(),
		)
		return nil
	}
}

func unaryMetricsInterceptor(service string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		started := time.Now()
		response, err := handler(ctx, req)
		code := status.Code(err).String()
		grpcServerRequests.WithLabelValues(service, info.FullMethod, code).Inc()
		grpcServerDuration.WithLabelValues(service, info.FullMethod).Observe(time.Since(started).Seconds())
		return response, err
	}
}

func recoveryInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response any, err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.ErrorContext(ctx, "grpc panic recovered",
					"method", info.FullMethod,
					"panic", recovered,
				)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
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
