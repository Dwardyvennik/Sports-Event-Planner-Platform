package grpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/university/sports-event-planner-platform/services/notification-service/internal/domain"
	"github.com/university/sports-event-planner-platform/services/notification-service/internal/usecase"
	notificationv1 "github.com/university/sports-event-planner-platform/services/notification-service/proto/notification/v1"
)

func Register(server *grpc.Server, notifications *usecase.NotificationUseCase, log *slog.Logger) {
	notificationv1.RegisterNotificationServiceServer(server, &Server{
		notifications: notifications,
		log:           log,
	})
	log.Info("notification grpc delivery registered")
}

type Server struct {
	notificationv1.UnimplementedNotificationServiceServer
	notifications *usecase.NotificationUseCase
	log           *slog.Logger
}

func (s *Server) SendNotification(ctx context.Context, req *notificationv1.SendNotificationRequest) (*notificationv1.NotificationResponse, error) {
	if strings.TrimSpace(req.UserId) == "" || strings.TrimSpace(req.Subject) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and subject are required")
	}

	input := usecase.SendNotificationInput{
		UserID:  req.UserId,
		Channel: req.Channel,
		Subject: req.Subject,
		Body:    req.Body,
	}

	if err := s.notifications.SendNotification(ctx, input); err != nil {
		s.log.Error("SendNotification rpc", "error", err)
		return nil, grpcError(err)
	}

	return &notificationv1.NotificationResponse{
		NotificationId: req.UserId,
		Status:         "sent",
	}, nil
}

func (s *Server) SendReminder(ctx context.Context, req *notificationv1.SendReminderRequest) (*notificationv1.NotificationResponse, error) {
	if strings.TrimSpace(req.EventId) == "" || strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id and user_id are required")
	}

	scheduledAt := time.Now().Add(time.Hour)
	if req.ScheduledAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.ScheduledAt)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "scheduled_at must be RFC3339: %v", err)
		}
		scheduledAt = parsed
	}

	input := usecase.SendReminderInput{
		EventID:     req.EventId,
		UserID:      req.UserId,
		Message:     req.Message,
		ScheduledAt: scheduledAt,
	}

	if err := s.notifications.SendReminder(ctx, input); err != nil {
		s.log.Error("SendReminder rpc", "error", err)
		return nil, grpcError(err)
	}

	return &notificationv1.NotificationResponse{
		NotificationId: req.EventId,
		Status:         "scheduled",
	}, nil
}

func (s *Server) GetNotifications(ctx context.Context, req *notificationv1.GetNotificationsRequest) (*notificationv1.GetNotificationsResponse, error) {
	if strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	notifications, err := s.notifications.GetNotifications(ctx, req.UserId)
	if err != nil {
		s.log.Error("GetNotifications rpc", "error", err)
		return nil, grpcError(err)
	}

	var pbNotifications []*notificationv1.Notification
	for _, n := range notifications {
		pbNotifications = append(pbNotifications, &notificationv1.Notification{
			Id:      n.ID,
			UserId:  n.UserID,
			Channel: n.Channel,
			Subject: n.Subject,
			Status:  n.Status,
		})
	}

	return &notificationv1.GetNotificationsResponse{
		Notifications: pbNotifications,
	}, nil
}

func grpcError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidNotification):
		return status.Error(codes.InvalidArgument, "invalid notification")
	case errors.Is(err, domain.ErrNotificationNotFound):
		return status.Error(codes.NotFound, "notification not found")
	default:
		return status.Error(codes.Internal, "notification service error")
	}
}
