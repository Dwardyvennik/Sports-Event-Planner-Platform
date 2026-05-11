package http

import (
	"context"
	"log/slog"
	nethttp "net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/university/sports-event-planner-platform/pkg/health"
	"github.com/university/sports-event-planner-platform/services/api-gateway/internal/delivery/grpc"
	authv1 "github.com/university/sports-event-planner-platform/services/auth-service/proto/auth/v1"
	eventv1 "github.com/university/sports-event-planner-platform/services/event-service/proto/event/v1"
	notificationv1 "github.com/university/sports-event-planner-platform/services/notification-service/proto/notification/v1"
)

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type eventRequest struct {
	Title       string `json:"title" binding:"required"`
	Sport       string `json:"sport" binding:"required"`
	Venue       string `json:"venue" binding:"required"`
	ScheduledAt string `json:"scheduled_at" binding:"required"`
	Capacity    int32  `json:"capacity" binding:"required,min=1"`
}

type notificationRequest struct {
	UserID  string `json:"user_id" binding:"required"`
	Channel string `json:"channel" binding:"required"`
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body" binding:"required"`
}

type reminderRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	Message     string `json:"message" binding:"required"`
	ScheduledAt string `json:"scheduled_at" binding:"required"`
}

func NewRouter(clients *grpc.Clients, checks map[string]health.Checker, log *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery(), LoggingMiddleware(log), PrometheusMiddleware())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(nethttp.StatusOK, gin.H{"service": "api-gateway", "status": "ok"})
	})
	router.GET("/readyz", readinessHandler(checks))
	router.GET("/metrics", MetricsHandler())

	v1 := router.Group("/v1")
	v1.POST("/auth/register", registerHandler(clients.Auth))
	v1.POST("/auth/login", loginHandler(clients.Auth))

	protected := v1.Group("")
	protected.Use(JWTMiddleware(clients.Auth))
	protected.GET("/auth/me", profileHandler(clients.Auth))
	protected.POST("/events", createEventHandler(clients.Event))
	protected.PUT("/events/:id", updateEventHandler(clients.Event))
	protected.GET("/events/:id", getEventHandler(clients.Event))
	protected.GET("/events", listEventsHandler(clients.Event))
	protected.POST("/events/:id/join", joinEventHandler(clients.Event))
	protected.DELETE("/events/:id/join", leaveEventHandler(clients.Event))
	protected.GET("/users/:id/events", userEventsHandler(clients.Event))
	protected.POST("/notifications", sendNotificationHandler(clients.Notification))
	protected.POST("/events/:id/reminders", sendReminderHandler(clients.Notification))
	protected.GET("/users/:id/notifications", userNotificationsHandler(clients.Notification))

	return router
}

func readinessHandler(checks map[string]health.Checker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		results := gin.H{}
		statusCode := nethttp.StatusOK
		for name, check := range checks {
			if err := check(ctx); err != nil {
				statusCode = nethttp.StatusServiceUnavailable
				results[name] = err.Error()
				continue
			}
			results[name] = "ok"
		}
		c.JSON(statusCode, gin.H{"service": "api-gateway", "status": nethttp.StatusText(statusCode), "checks": results})
	}
}

func registerHandler(client authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req registerRequest
		if !bindJSON(c, &req) {
			return
		}

		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.Register(ctx, &authv1.RegisterRequest{
			Email:    req.Email,
			Password: req.Password,
			Role:     req.Role,
		})
		respond(c, resp, err)
	}
}

func loginHandler(client authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if !bindJSON(c, &req) {
			return
		}

		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.Login(ctx, &authv1.LoginRequest{Email: req.Email, Password: req.Password})
		respond(c, resp, err)
	}
}

func profileHandler(client authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.GetProfile(ctx, &authv1.GetProfileRequest{UserId: currentUserID(c)})
		respond(c, resp, err)
	}
}

func createEventHandler(client eventv1.EventServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req eventRequest
		if !bindJSON(c, &req) {
			return
		}

		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.CreateEvent(ctx, &eventv1.CreateEventRequest{
			Title:       req.Title,
			Sport:       req.Sport,
			Venue:       req.Venue,
			ScheduledAt: req.ScheduledAt,
			Capacity:    req.Capacity,
		})
		respond(c, resp, err)
	}
}

func updateEventHandler(client eventv1.EventServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req eventRequest
		if !bindJSON(c, &req) {
			return
		}

		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.UpdateEvent(ctx, &eventv1.UpdateEventRequest{
			EventId:     c.Param("id"),
			Title:       req.Title,
			Sport:       req.Sport,
			Venue:       req.Venue,
			ScheduledAt: req.ScheduledAt,
			Capacity:    req.Capacity,
		})
		respond(c, resp, err)
	}
}

func getEventHandler(client eventv1.EventServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.GetEvent(ctx, &eventv1.GetEventRequest{EventId: c.Param("id")})
		respond(c, resp, err)
	}
}

func listEventsHandler(client eventv1.EventServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.ListEvents(ctx, &eventv1.ListEventsRequest{
			Sport:    c.Query("sport"),
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		respond(c, resp, err)
	}
}

func joinEventHandler(client eventv1.EventServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.JoinEvent(ctx, &eventv1.JoinEventRequest{EventId: c.Param("id"), UserId: currentUserID(c)})
		respond(c, resp, err)
	}
}

func leaveEventHandler(client eventv1.EventServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.LeaveEvent(ctx, &eventv1.LeaveEventRequest{EventId: c.Param("id"), UserId: currentUserID(c)})
		respond(c, resp, err)
	}
}

func userEventsHandler(client eventv1.EventServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.GetUserEvents(ctx, &eventv1.GetUserEventsRequest{UserId: c.Param("id")})
		respond(c, resp, err)
	}
}

func sendNotificationHandler(client notificationv1.NotificationServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req notificationRequest
		if !bindJSON(c, &req) {
			return
		}

		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.SendNotification(ctx, &notificationv1.SendNotificationRequest{
			UserId:  req.UserID,
			Channel: req.Channel,
			Subject: req.Subject,
			Body:    req.Body,
		})
		respond(c, resp, err)
	}
}

func sendReminderHandler(client notificationv1.NotificationServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req reminderRequest
		if !bindJSON(c, &req) {
			return
		}

		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.SendReminder(ctx, &notificationv1.SendReminderRequest{
			EventId:     c.Param("id"),
			UserId:      req.UserID,
			Message:     req.Message,
			ScheduledAt: req.ScheduledAt,
		})
		respond(c, resp, err)
	}
}

func userNotificationsHandler(client notificationv1.NotificationServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := timeoutContext(c)
		defer cancel()

		resp, err := client.GetNotifications(ctx, &notificationv1.GetNotificationsRequest{UserId: c.Param("id")})
		respond(c, resp, err)
	}
}

func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}

func respond(c *gin.Context, payload any, err error) {
	if err != nil {
		c.JSON(statusCodeFromGRPC(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(nethttp.StatusOK, payload)
}

func statusCodeFromGRPC(err error) int {
	code := status.Code(err)
	switch code {
	case codes.InvalidArgument:
		return nethttp.StatusBadRequest
	case codes.Unauthenticated:
		return nethttp.StatusUnauthorized
	case codes.PermissionDenied:
		return nethttp.StatusForbidden
	case codes.NotFound:
		return nethttp.StatusNotFound
	case codes.Unavailable, codes.DeadlineExceeded:
		return nethttp.StatusServiceUnavailable
	default:
		return nethttp.StatusBadGateway
	}
}

func timeoutContext(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), 3*time.Second)
}

func currentUserID(c *gin.Context) string {
	value, ok := c.Get("user_id")
	if !ok {
		return ""
	}
	userID, _ := value.(string)
	return userID
}
