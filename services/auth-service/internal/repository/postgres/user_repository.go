package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
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
