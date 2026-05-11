package grpc

import (
	"context"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	if err := s.notifications.Health(ctx); err != nil {
		return nil, status.Error(codes.Unavailable, "notification dependencies unavailable")
	}
	return &notificationv1.NotificationResponse{NotificationId: "notification_stub", Status: "queued"}, nil
}

func (s *Server) SendReminder(ctx context.Context, req *notificationv1.SendReminderRequest) (*notificationv1.NotificationResponse, error) {
	if strings.TrimSpace(req.EventId) == "" || strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id and user_id are required")
	}
	if err := s.notifications.Health(ctx); err != nil {
		return nil, status.Error(codes.Unavailable, "notification dependencies unavailable")
	}
	return &notificationv1.NotificationResponse{NotificationId: "reminder_stub", Status: "scheduled"}, nil
}

func (s *Server) GetNotifications(ctx context.Context, req *notificationv1.GetNotificationsRequest) (*notificationv1.GetNotificationsResponse, error) {
	if strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if err := s.notifications.Health(ctx); err != nil {
		return nil, status.Error(codes.Unavailable, "notification dependencies unavailable")
	}
	return &notificationv1.GetNotificationsResponse{Notifications: []*notificationv1.Notification{
		{Id: "notification_1", UserId: req.UserId, Channel: "email", Subject: "Event update", Status: "queued"},
	}}, nil
}
