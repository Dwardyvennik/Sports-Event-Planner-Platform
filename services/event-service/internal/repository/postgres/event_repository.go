package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/university/sports-event-planner-platform/services/event-service/internal/domain"
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

func (r *EventRepository) Create(ctx context.Context, event *domain.Event) error {
	const query = `
		INSERT INTO events (
			sport,
			competition,
			category,
			title,
			description,
			start_time,
			end_time,
			status,
			country,
			city
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id::text, created_at, updated_at`

	return r.pool.QueryRow(
		ctx,
		query,
		event.Sport,
		event.Competition,
		event.Category,
		event.Title,
		event.Description,
		event.StartTime,
		nullableTime(event.EndTime),
		event.Status,
		event.Country,
		event.City,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
}

func (r *EventRepository) Get(ctx context.Context, id string) (*domain.Event, error) {
	const query = `
		SELECT
			id::text,
			sport,
			COALESCE(competition, ''),
			COALESCE(category, ''),
			title,
			COALESCE(description, ''),
			start_time,
			end_time,
			status,
			COALESCE(country, ''),
			COALESCE(city, ''),
			created_at,
			updated_at
		FROM events
		WHERE id = $1::uuid`

	event := new(domain.Event)
	var endTime *time.Time
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&event.ID,
		&event.Sport,
		&event.Competition,
		&event.Category,
		&event.Title,
		&event.Description,
		&event.StartTime,
		&endTime,
		&event.Status,
		&event.Country,
		&event.City,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrEventNotFound
	}
	if err != nil {
		return nil, err
	}
	if endTime != nil {
		event.EndTime = *endTime
	}
	return event, nil
}

func (r *EventRepository) UpdateEvent(ctx context.Context, id string, title string, sport string, location string, startTime time.Time, endTime time.Time, maxParticipants int32) (*domain.Event, error) {
	const query = `
		UPDATE events
		SET title = $2,
			sport = $3,
			location = $4,
			start_time = $5,
			end_time = $6,
			max_participants = $7,
			updated_at = now()
		WHERE id = $1::uuid
		RETURNING
			id::text,
			sport,
			COALESCE(competition, ''),
			COALESCE(category, ''),
			title,
			COALESCE(description, ''),
			start_time,
			end_time,
			status,
			COALESCE(country, ''),
			COALESCE(city, ''),
			created_at,
			updated_at`

	event := new(domain.Event)
	var returnedEndTime *time.Time
	err := r.pool.QueryRow(ctx, query, id, title, sport, location, startTime, nullableTime(endTime), maxParticipants).Scan(
		&event.ID,
		&event.Sport,
		&event.Competition,
		&event.Category,
		&event.Title,
		&event.Description,
		&event.StartTime,
		&returnedEndTime,
		&event.Status,
		&event.Country,
		&event.City,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrEventNotFound
	}
	if err != nil {
		return nil, err
	}
	if returnedEndTime != nil {
		event.EndTime = *returnedEndTime
	}
	return event, nil
}

func (r *EventRepository) List(ctx context.Context, filter domain.EventFilter) ([]*domain.Event, error) {
	query, args := listQuery(filter)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []*domain.Event{}
	for rows.Next() {
		event := new(domain.Event)
		var endTime *time.Time
		if err := rows.Scan(
			&event.ID,
			&event.Sport,
			&event.Competition,
			&event.Category,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&endTime,
			&event.Status,
			&event.Country,
			&event.City,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if endTime != nil {
			event.EndTime = *endTime
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM events WHERE id = $1::uuid`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if isInvalidUUID(err) {
		return domain.ErrEventNotFound
	}
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrEventNotFound
	}
	return nil
}

func listQuery(filter domain.EventFilter) (string, []any) {
	var builder strings.Builder
	builder.WriteString(`
		SELECT
			id::text,
			sport,
			COALESCE(competition, ''),
			COALESCE(category, ''),
			title,
			COALESCE(description, ''),
			start_time,
			end_time,
			status,
			COALESCE(country, ''),
			COALESCE(city, ''),
			created_at,
			updated_at
		FROM events`)

	args := make([]any, 0, 7)
	conditions := make([]string, 0, 5)
	addCondition := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}

	if filter.Sport != "" {
		addCondition("sport = $%d", filter.Sport)
	}
	if filter.Competition != "" {
		addCondition("competition = $%d", filter.Competition)
	}
	if !filter.StartTimeFrom.IsZero() {
		addCondition("start_time >= $%d", filter.StartTimeFrom)
	}
	if !filter.StartTimeTo.IsZero() {
		addCondition("start_time <= $%d", filter.StartTimeTo)
	}
	if filter.Country != "" {
		addCondition("country = $%d", filter.Country)
	}

	if len(conditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(conditions, " AND "))
	}

	args = append(args, filter.Limit)
	limitIndex := len(args)
	args = append(args, filter.Offset)
	offsetIndex := len(args)
	builder.WriteString(fmt.Sprintf(" ORDER BY start_time ASC, id ASC LIMIT $%d OFFSET $%d", limitIndex, offsetIndex))

	return builder.String(), args
}

func nullableTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}

func isInvalidUUID(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}
