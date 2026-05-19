package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/university/sports-event-planner-platform/services/auth-service/internal/domain"
)

type UserRepository interface {
	Ping(context.Context) error
	CreateUser(context.Context, *domain.User) error
	FindByEmail(context.Context, string) (*domain.User, error)
	FindByID(context.Context, string) (*domain.User, error)
	CreateSession(context.Context, *domain.Session) error
	FindSession(context.Context, string) (*domain.Session, error)
}

type Config struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type AuthTokens struct {
	UserID       string
	Email        string
	Role         string
	AccessToken  string
	RefreshToken string
}

type AuthUseCase struct {
	users           UserRepository
	cache           *redis.Client
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewAuthUseCase(users UserRepository, args ...any) *AuthUseCase {
	var cache *redis.Client
	var cfg Config
	switch len(args) {
	case 0:
	case 1:
		if value, ok := args[0].(Config); ok {
			cfg = value
		}
	default:
		if value, ok := args[0].(*redis.Client); ok {
			cache = value
		}
		if len(args) > 1 {
			if value, ok := args[1].(Config); ok {
				cfg = value
			}
		}
	}
	if cfg.AccessTokenTTL <= 0 {
		cfg.AccessTokenTTL = time.Hour
	}
	if cfg.RefreshTokenTTL <= 0 {
		cfg.RefreshTokenTTL = 7 * 24 * time.Hour
	}
	return &AuthUseCase{
		users:           users,
		cache:           cache,
		jwtSecret:       []byte(cfg.JWTSecret),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
	}
}

func (u *AuthUseCase) Health(ctx context.Context) error {
	if u.users == nil {
		return nil
	}
	return u.users.Ping(ctx)
}

func (u *AuthUseCase) Register(ctx context.Context, email, password, role string) (*AuthTokens, error) {
	email = normalizeEmail(email)
	role = normalizeRole(role)
	if email == "" || password == "" {
		return nil, domain.ErrInvalidCredentials
	}
	if len(password) < 8 {
		return nil, domain.ErrInvalidCredentials
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         role,
	}
	if err := u.users.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return u.issueTokens(ctx, user)
}

func (u *AuthUseCase) Login(ctx context.Context, email, password string) (*AuthTokens, error) {
	email = normalizeEmail(email)
	if email == "" || password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := u.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		return nil, domain.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return u.issueTokens(ctx, user)
}

func (u *AuthUseCase) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	claims, err := u.parseToken(token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := u.users.FindByID(ctx, claims.UserID)
	if errors.Is(err, domain.ErrUserNotFound) {
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *AuthUseCase) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domain.ErrUserNotFound
	}
	if user, ok := u.profileFromCache(ctx, userID); ok {
		return user, nil
	}
	user, err := u.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	u.cacheProfile(ctx, user)
	return user, nil
}

func (u *AuthUseCase) RefreshToken(ctx context.Context, rawRefreshToken string) (*AuthTokens, error) {
	rawRefreshToken = strings.TrimSpace(rawRefreshToken)
	if rawRefreshToken == "" {
		return nil, domain.ErrInvalidToken
	}

	session, err := u.users.FindSession(ctx, hashToken(rawRefreshToken))
	if errors.Is(err, domain.ErrInvalidToken) {
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	if session == nil || time.Now().UTC().After(session.ExpiresAt) {
		return nil, domain.ErrInvalidToken
	}

	user, err := u.users.FindByID(ctx, session.UserID)
	if errors.Is(err, domain.ErrUserNotFound) {
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}

	return u.issueTokens(ctx, user)
}

func (u *AuthUseCase) profileFromCache(ctx context.Context, id string) (*domain.User, bool) {
	if u.cache == nil {
		return nil, false
	}
	data, err := u.cache.Get(ctx, "user:"+id).Bytes()
	if err != nil {
		return nil, false
	}
	user := new(domain.User)
	if err := json.Unmarshal(data, user); err != nil {
		return nil, false
	}
	return user, true
}

func (u *AuthUseCase) cacheProfile(ctx context.Context, user *domain.User) {
	if u.cache == nil || user == nil || user.ID == "" {
		return
	}
	data, err := json.Marshal(user)
	if err != nil {
		return
	}
	_ = u.cache.Set(ctx, "user:"+user.ID, data, 5*time.Minute).Err()
}

func (u *AuthUseCase) issueTokens(ctx context.Context, user *domain.User) (*AuthTokens, error) {
	accessToken, err := u.signToken(user, u.accessTokenTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, err := u.signToken(user, u.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		UserID:           user.ID,
		RefreshTokenHash: hashToken(refreshToken),
		ExpiresAt:        time.Now().UTC().Add(u.refreshTokenTTL),
	}
	if err := u.users.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return &AuthTokens{
		UserID:       user.ID,
		Email:        user.Email,
		Role:         user.Role,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *AuthUseCase) signToken(user *domain.User, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	tokenID, err := newTokenID()
	if err != nil {
		return "", err
	}
	claims := authClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ID:        tokenID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(u.jwtSecret)
}

func (u *AuthUseCase) parseToken(rawToken string) (*authClaims, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, domain.ErrInvalidToken
	}

	claims := new(authClaims)
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		return u.jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid || claims.UserID == "" || claims.Email == "" || claims.Role == "" {
		return nil, domain.ErrInvalidToken
	}
	return claims, nil
}

type authClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeRole(role string) string {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		return "student"
	}
	return role
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func newTokenID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
