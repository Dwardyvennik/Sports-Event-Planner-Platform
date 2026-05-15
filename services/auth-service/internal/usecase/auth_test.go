package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/university/sports-event-planner-platform/services/auth-service/internal/domain"
)

func TestRegisterHashesPasswordAndIssuesValidToken(t *testing.T) {
	repo := newFakeUserRepository()
	auth := NewAuthUseCase(repo, Config{
		JWTSecret:       "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	})

	tokens, err := auth.Register(context.Background(), " Student@Example.COM ", "password123", "")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if tokens.UserID == "" || tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected issued tokens and user id, got %#v", tokens)
	}

	user := repo.usersByID[tokens.UserID]
	if user.Email != "student@example.com" {
		t.Fatalf("email was not normalized: %q", user.Email)
	}
	if user.Role != "student" {
		t.Fatalf("expected default role student, got %q", user.Role)
	}
	if user.PasswordHash == "password123" {
		t.Fatal("password was stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")); err != nil {
		t.Fatalf("stored hash does not match password: %v", err)
	}
	if len(repo.sessions) != 1 {
		t.Fatalf("expected one refresh session, got %d", len(repo.sessions))
	}

	validated, err := auth.ValidateToken(context.Background(), tokens.AccessToken)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if validated.ID != tokens.UserID || validated.Email != "student@example.com" || validated.Role != "student" {
		t.Fatalf("unexpected validated user: %#v", validated)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	repo := newFakeUserRepository()
	auth := NewAuthUseCase(repo, Config{JWTSecret: "test-secret"})

	if _, err := auth.Register(context.Background(), "student@example.com", "password123", "student"); err != nil {
		t.Fatalf("first register: %v", err)
	}
	_, err := auth.Register(context.Background(), "student@example.com", "password123", "student")
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	repo := newFakeUserRepository()
	auth := NewAuthUseCase(repo, Config{JWTSecret: "test-secret"})

	if _, err := auth.Register(context.Background(), "student@example.com", "password123", "student"); err != nil {
		t.Fatalf("register: %v", err)
	}
	_, err := auth.Login(context.Background(), "student@example.com", "wrong-password")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestValidateTokenRejectsInvalidToken(t *testing.T) {
	auth := NewAuthUseCase(newFakeUserRepository(), Config{JWTSecret: "test-secret"})

	_, err := auth.ValidateToken(context.Background(), "not-a-jwt")
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
}

type fakeUserRepository struct {
	usersByID    map[string]*domain.User
	usersByEmail map[string]*domain.User
	sessions     map[string]*domain.Session
	nextID       int
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		usersByID:    map[string]*domain.User{},
		usersByEmail: map[string]*domain.User{},
		sessions:     map[string]*domain.Session{},
		nextID:       1,
	}
}

func (r *fakeUserRepository) Ping(context.Context) error {
	return nil
}

func (r *fakeUserRepository) CreateUser(_ context.Context, user *domain.User) error {
	if _, exists := r.usersByEmail[user.Email]; exists {
		return domain.ErrUserAlreadyExists
	}
	created := *user
	created.ID = fmt.Sprintf("00000000-0000-0000-0000-%012d", r.nextID)
	created.CreatedAt = time.Now().UTC()
	created.UpdatedAt = created.CreatedAt
	r.nextID++

	r.usersByID[created.ID] = &created
	r.usersByEmail[created.Email] = &created
	user.ID = created.ID
	user.CreatedAt = created.CreatedAt
	user.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *fakeUserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	user, ok := r.usersByEmail[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	copy := *user
	return &copy, nil
}

func (r *fakeUserRepository) FindByID(_ context.Context, id string) (*domain.User, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	copy := *user
	return &copy, nil
}

func (r *fakeUserRepository) CreateSession(_ context.Context, session *domain.Session) error {
	created := *session
	created.ID = fmt.Sprintf("session-%d", len(r.sessions)+1)
	created.CreatedAt = time.Now().UTC()
	r.sessions[created.RefreshTokenHash] = &created
	session.ID = created.ID
	session.CreatedAt = created.CreatedAt
	return nil
}

func (r *fakeUserRepository) FindSession(_ context.Context, refreshTokenHash string) (*domain.Session, error) {
	session, ok := r.sessions[refreshTokenHash]
	if !ok {
		return nil, domain.ErrInvalidToken
	}
	copy := *session
	return &copy, nil
}
