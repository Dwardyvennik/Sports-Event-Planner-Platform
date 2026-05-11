package grpc

import (
	"log/slog"

	"google.golang.org/grpc"

	"github.com/university/sports-event-planner-platform/services/event-service/internal/usecase"
)

func Register(server *grpc.Server, events *usecase.EventUseCase, log *slog.Logger) {
	_ = server
	_ = events
	log.Info("event grpc delivery registered")
}
