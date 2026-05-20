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

	"github.com/dwardyvennik/sports-event-planner-platform/services/event-service/internal/domain"
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
			creator_id,
			sport,
			competition,
			category,
			title,
			description,
			location,
			start_time,
			end_time,
			status,
			country,
			city,
			max_participants
		)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id::text, created_at, updated_at`

	return r.pool.QueryRow(
		ctx,
		query,
		event.CreatorID,
		event.Sport,
		event.Competition,
		event.Category,
		event.Title,
		event.Description,
		event.Location,
		event.StartTime,
		nullableTime(event.EndTime),
		event.Status,
		event.Country,
		event.City,
		event.MaxParticipants,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
}

func (r *EventRepository) Get(ctx context.Context, id string) (*domain.Event, error) {
	const query = eventSelect + ` WHERE e.id = $1::uuid GROUP BY e.id`

	event := new(domain.Event)
	err := scanEvent(r.pool.QueryRow(ctx, query, id), event)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return nil, domain.ErrEventNotFound
	}
	if err != nil {
		return nil, err
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
		if err := scanEvent(rows, event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepository) Update(ctx context.Context, event *domain.Event) error {
	if event.MaxParticipants > 0 {
		var participantsCount int32
		err := r.pool.QueryRow(ctx, `
			SELECT COUNT(*)::int
			FROM event_participants
			WHERE event_id = $1::uuid`, event.ID).Scan(&participantsCount)
		if isInvalidUUID(err) {
			return domain.ErrEventNotFound
		}
		if err != nil {
			return err
		}
		if participantsCount > event.MaxParticipants {
			return domain.ErrEventFull
		}
	}

	const query = `
		UPDATE events
		SET sport = $3,
			competition = $4,
			category = $5,
			title = $6,
			description = $7,
			location = $8,
			start_time = $9,
			end_time = $10,
			status = $11,
			country = $12,
			city = $13,
			max_participants = $14,
			updated_at = now()
		WHERE id = $1::uuid
			AND creator_id = $2::uuid
		RETURNING updated_at`

	err := r.pool.QueryRow(
		ctx,
		query,
		event.ID,
		event.CreatorID,
		event.Sport,
		event.Competition,
		event.Category,
		event.Title,
		event.Description,
		event.Location,
		event.StartTime,
		nullableTime(event.EndTime),
		event.Status,
		event.Country,
		event.City,
		event.MaxParticipants,
	).Scan(&event.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return r.notFoundOrForbidden(ctx, event.ID, event.CreatorID)
	}
	if isInvalidUUID(err) {
		return domain.ErrEventNotFound
	}
	return err
}

func (r *EventRepository) Delete(ctx context.Context, id string, userID string) error {
	const query = `DELETE FROM events WHERE id = $1::uuid AND creator_id = $2::uuid`

	commandTag, err := r.pool.Exec(ctx, query, id, userID)
	if isInvalidUUID(err) {
		return domain.ErrEventNotFound
	}
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return r.notFoundOrForbidden(ctx, id, userID)
	}
	return nil
}

func (r *EventRepository) Join(ctx context.Context, eventID string, userID string) (*domain.Event, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := lockAndValidateCapacity(ctx, tx, eventID, userID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO event_participants (event_id, user_id)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT DO NOTHING`, eventID, userID); err != nil {
		if isInvalidUUID(err) {
			return nil, domain.ErrEventNotFound
		}
		return nil, err
	}

	event := new(domain.Event)
	if err := scanEvent(tx.QueryRow(ctx, eventSelect+` WHERE e.id = $1::uuid GROUP BY e.id`, eventID), event); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return event, nil
}

func (r *EventRepository) Leave(ctx context.Context, eventID string, userID string) (*domain.Event, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM events WHERE id = $1::uuid)`, eventID).Scan(&exists); err != nil {
		if isInvalidUUID(err) {
			return nil, domain.ErrEventNotFound
		}
		return nil, err
	}
	if !exists {
		return nil, domain.ErrEventNotFound
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM event_participants
		WHERE event_id = $1::uuid AND user_id = $2::uuid`, eventID, userID); err != nil {
		if isInvalidUUID(err) {
			return nil, domain.ErrEventNotFound
		}
		return nil, err
	}

	event := new(domain.Event)
	if err := scanEvent(tx.QueryRow(ctx, eventSelect+` WHERE e.id = $1::uuid GROUP BY e.id`, eventID), event); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return event, nil
}

func (r *EventRepository) notFoundOrForbidden(ctx context.Context, eventID string, userID string) error {
	var ownerID string
	err := r.pool.QueryRow(ctx, `SELECT creator_id::text FROM events WHERE id = $1::uuid`, eventID).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return domain.ErrEventNotFound
	}
	if err != nil {
		return err
	}
	if ownerID != userID {
		return domain.ErrEventForbidden
	}
	return domain.ErrEventNotFound
}

func lockAndValidateCapacity(ctx context.Context, tx pgx.Tx, eventID string, userID string) error {
	var maxParticipants int32
	err := tx.QueryRow(ctx, `
		SELECT max_participants
		FROM events
		WHERE id = $1::uuid
		FOR UPDATE`, eventID).Scan(&maxParticipants)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return domain.ErrEventNotFound
	}
	if err != nil {
		return err
	}

	var alreadyJoined bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM event_participants
			WHERE event_id = $1::uuid AND user_id = $2::uuid
		)`, eventID, userID).Scan(&alreadyJoined); err != nil {
		return err
	}
	if alreadyJoined {
		return nil
	}

	var participantsCount int32
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM event_participants
		WHERE event_id = $1::uuid`, eventID).Scan(&participantsCount); err != nil {
		return err
	}
	if maxParticipants > 0 && participantsCount >= maxParticipants {
		return domain.ErrEventFull
	}
	return nil
}

const eventSelect = `
	SELECT
		e.id::text,
		e.creator_id::text,
		e.sport,
		COALESCE(e.competition, ''),
		COALESCE(e.category, ''),
		e.title,
		COALESCE(e.description, ''),
		COALESCE(e.location, ''),
		e.start_time,
		e.end_time,
		e.status,
		COALESCE(e.country, ''),
		COALESCE(e.city, ''),
		e.max_participants,
		COUNT(ep.user_id)::int,
		e.created_at,
		e.updated_at
	FROM events e
	LEFT JOIN event_participants ep ON ep.event_id = e.id`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanEvent(row rowScanner, event *domain.Event) error {
	var endTime *time.Time
	err := row.Scan(
		&event.ID,
		&event.CreatorID,
		&event.Sport,
		&event.Competition,
		&event.Category,
		&event.Title,
		&event.Description,
		&event.Location,
		&event.StartTime,
		&endTime,
		&event.Status,
		&event.Country,
		&event.City,
		&event.MaxParticipants,
		&event.ParticipantsCount,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if endTime != nil {
		event.EndTime = *endTime
	}
	return nil
}

func listQuery(filter domain.EventFilter) (string, []any) {
	var builder strings.Builder
	builder.WriteString(eventSelect)

	args := make([]any, 0, 9)
	conditions := make([]string, 0, 7)
	addCondition := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}

	if filter.Sport != "" {
		addCondition("e.sport = $%d", filter.Sport)
	}
	if filter.Category != "" {
		addCondition("e.category = $%d", filter.Category)
	}
	if filter.Competition != "" {
		addCondition("e.competition = $%d", filter.Competition)
	}
	if filter.Status != "" {
		addCondition("e.status = $%d", filter.Status)
	}
	if !filter.StartTimeFrom.IsZero() {
		addCondition("e.start_time >= $%d", filter.StartTimeFrom)
	}
	if !filter.StartTimeTo.IsZero() {
		addCondition("e.start_time <= $%d", filter.StartTimeTo)
	}
	if filter.Country != "" {
		addCondition("e.country = $%d", filter.Country)
	}
	if filter.City != "" {
		addCondition("e.city = $%d", filter.City)
	}

	if len(conditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(conditions, " AND "))
	}

	builder.WriteString(" GROUP BY e.id")
	args = append(args, filter.Limit)
	limitIndex := len(args)
	args = append(args, filter.Offset)
	offsetIndex := len(args)
	builder.WriteString(fmt.Sprintf(" ORDER BY e.start_time ASC, e.id ASC LIMIT $%d OFFSET $%d", limitIndex, offsetIndex))

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
