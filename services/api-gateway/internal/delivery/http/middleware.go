package http

import (
	"log/slog"
	nethttp "net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwardyvennik/sports-event-planner-platform/pkg/metrics"
	authv1 "github.com/dwardyvennik/sports-event-planner-platform/services/auth-service/proto/auth/v1"
)

const serviceName = "api-gateway"

func MetricsHandler() gin.HandlerFunc {
	handler := metrics.Handler(serviceName)
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

		metrics.ObserveHTTP(serviceName, c.Request.Method, path, c.Writer.Status(), time.Since(started))
	}
}

func JWTMiddleware(auth authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeAPIError(c, nethttp.StatusUnauthorized, errorUnauthorized, "missing bearer token")
			c.Abort()
			return
		}

		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := auth.ValidateToken(ctx, &authv1.ValidateTokenRequest{
			Token: strings.TrimPrefix(header, "Bearer "),
		})
		if err != nil || !resp.Valid {
			writeAPIError(c, nethttp.StatusUnauthorized, errorInvalidToken, "invalid token")
			c.Abort()
			return
		}

		c.Set("user_id", resp.UserId)
		c.Set("role", resp.Role)
		c.Next()
	}
}
