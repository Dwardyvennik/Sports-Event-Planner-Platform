package grpc

import (
	"log/slog"

	"google.golang.org/grpc"

	"github.com/university/sports-event-planner-platform/services/auth-service/internal/usecase"
)

func Register(server *grpc.Server, auth *usecase.AuthUseCase, log *slog.Logger) {
	_ = server
	_ = auth
	log.Info("auth grpc delivery registered")
}
