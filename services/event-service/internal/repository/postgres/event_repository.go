package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

func (r *EventRepository) Ping(ctx context.Context) error {
	if r.pool == nil {
		return nil
	}
	return r.pool.Ping(ctx)
}
