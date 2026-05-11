package http

import (
	"log/slog"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	authv1 "github.com/university/sports-event-planner-platform/services/auth-service/proto/auth/v1"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "sports_planner",
		Subsystem: "api_gateway",
		Name:      "http_requests_total",
		Help:      "Total number of API gateway HTTP requests.",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "sports_planner",
		Subsystem: "api_gateway",
		Name:      "http_request_duration_seconds",
		Help:      "API gateway HTTP request duration.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path"})
)

func MetricsHandler() gin.HandlerFunc {
	handler := promhttp.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

func LoggingMiddleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		log.InfoContext(c.Request.Context(), "gateway request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"duration_ms", time.Since(started).Milliseconds(),
		)
	}
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, nethttp.StatusText(c.Writer.Status())).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(time.Since(started).Seconds())
	}
}

func JWTMiddleware(auth authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(nethttp.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := auth.ValidateToken(ctx, &authv1.ValidateTokenRequest{
			Token: strings.TrimPrefix(header, "Bearer "),
		})
		if err != nil || !resp.Valid {
			c.AbortWithStatusJSON(nethttp.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user_id", resp.UserId)
		c.Set("role", resp.Role)
		c.Next()
	}
}
