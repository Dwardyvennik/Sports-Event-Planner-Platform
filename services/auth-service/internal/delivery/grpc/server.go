package grpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/university/sports-event-planner-platform/services/auth-service/internal/domain"
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
	tokens, err := s.auth.Register(ctx, req.Email, req.Password, req.Role)
	if err != nil {
		return nil, grpcError(err)
	}
	return &authv1.AuthResponse{
		UserId:       tokens.UserID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}
	tokens, err := s.auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, grpcError(err)
	}
	return &authv1.AuthResponse{
		UserId:       tokens.UserID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	user, err := s.auth.ValidateToken(ctx, req.Token)
	if errors.Is(err, domain.ErrInvalidToken) {
		return &authv1.ValidateTokenResponse{Valid: false}, nil
	}
	if err != nil {
		return nil, grpcError(err)
	}
	return &authv1.ValidateTokenResponse{
		Valid:  true,
		UserId: user.ID,
		Role:   user.Role,
	}, nil
}

func (s *Server) GetProfile(ctx context.Context, req *authv1.GetProfileRequest) (*authv1.UserProfileResponse, error) {
	if strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	user, err := s.auth.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, grpcError(err)
	}
	return &authv1.UserProfileResponse{
		UserId: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}, nil
}

func grpcError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "invalid token")
	default:
		return status.Error(codes.Internal, "auth service error")
	}
}
