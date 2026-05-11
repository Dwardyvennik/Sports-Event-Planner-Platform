package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) Ping(ctx context.Context) error {
	if r.pool == nil {
		return nil
	}
	return r.pool.Ping(ctx)
}
