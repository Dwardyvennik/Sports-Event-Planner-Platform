package grpc

import (
	"log/slog"

	"google.golang.org/grpc"

	"github.com/university/sports-event-planner-platform/services/notification-service/internal/usecase"
)

func Register(server *grpc.Server, notifications *usecase.NotificationUseCase, log *slog.Logger) {
	_ = server
	_ = notifications
	log.Info("notification grpc delivery registered")
}
