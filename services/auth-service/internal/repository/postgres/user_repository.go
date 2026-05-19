package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/university/sports-event-planner-platform/services/auth-service/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Ping(ctx context.Context) error {
	if r.pool == nil {
		return nil
	}
	return r.pool.Ping(ctx)
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id::text, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, user.Email, user.PasswordHash, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if isUniqueViolation(err) {
		return domain.ErrUserAlreadyExists
	}
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id::text, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1`

	user := new(domain.User)
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	const query = `
		SELECT id::text, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1::uuid`

	user := new(domain.User)
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	const query = `
		INSERT INTO sessions (user_id, refresh_token_hash, expires_at)
		VALUES ($1::uuid, $2, $3)
		RETURNING id::text, created_at`

	return r.pool.QueryRow(ctx, query, session.UserID, session.RefreshTokenHash, session.ExpiresAt).
		Scan(&session.ID, &session.CreatedAt)
}

func (r *UserRepository) FindSession(ctx context.Context, refreshTokenHash string) (*domain.Session, error) {
	const query = `
		SELECT id::text, user_id::text, refresh_token_hash, expires_at, created_at
		FROM sessions
		WHERE refresh_token_hash = $1
			AND expires_at > now()`

	session := new(domain.Session)
	err := r.pool.QueryRow(ctx, query, refreshTokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	return session, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
