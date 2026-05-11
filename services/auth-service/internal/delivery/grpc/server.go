package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/university/sports-event-planner-platform/services/auth-service/internal/usecase"
	authv1 "github.com/university/sports-event-planner-platform/services/auth-service/proto/auth/v1"
)

func Register(server *grpc.Server, auth *usecase.AuthUseCase, log *slog.Logger) {
	authv1.RegisterAuthServiceServer(server, &Server{
		auth: auth,
		log:  log,
	})
	log.Info("auth grpc delivery registered")
}

type Server struct {
	authv1.UnimplementedAuthServiceServer
	auth *usecase.AuthUseCase
	log  *slog.Logger
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	if err := s.auth.Health(ctx); err != nil {
		return nil, status.Error(codes.Unavailable, "auth dependencies unavailable")
	}
	userID := fmt.Sprintf("user_%x", len(req.Email)+len(req.Role))
	return &authv1.AuthResponse{
		UserId:       userID,
		AccessToken:  "dev-access-token-" + userID,
		RefreshToken: "dev-refresh-token-" + userID,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	if err := s.auth.Health(ctx); err != nil {
		return nil, status.Error(codes.Unavailable, "auth dependencies unavailable")
	}
	userID := fmt.Sprintf("user_%x", len(req.Email))
	return &authv1.AuthResponse{
		UserId:       userID,
		AccessToken:  "dev-access-token-" + userID,
		RefreshToken: "dev-refresh-token-" + userID,
	}, nil
}

func (s *Server) ValidateToken(_ context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	token := strings.TrimSpace(req.Token)
	if token == "" {
		return &authv1.ValidateTokenResponse{Valid: false}, nil
	}
	return &authv1.ValidateTokenResponse{
		Valid:  true,
		UserId: "user_from_token",
		Role:   "student",
	}, nil
}

func (s *Server) GetProfile(_ context.Context, req *authv1.GetProfileRequest) (*authv1.UserProfileResponse, error) {
	if strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	return &authv1.UserProfileResponse{
		UserId: req.UserId,
		Email:  req.UserId + "@university.example",
		Role:   "student",
	}, nil
}
