package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/university/sports-event-planner-platform/services/notification-service/internal/domain"
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


func (r *NotificationRepository) SaveNotification(ctx context.Context, n *domain.Notification) error {
	query := `
		INSERT INTO notifications (user_id, channel, subject, body, status, sent_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		n.UserID,
		n.Channel,
		n.Subject,
		n.Body,
		n.Status,
		n.SentAt,
	).Scan(&n.ID, &n.CreatedAt)
}


func (r *NotificationRepository) GetUserNotifications(ctx context.Context, userID string) ([]*domain.Notification, error) {
	query := `
		SELECT id, user_id, channel, subject, body, status, created_at, sent_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		n := &domain.Notification{}
		if err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Channel,
			&n.Subject,
			&n.Body,
			&n.Status,
			&n.CreatedAt,
			&n.SentAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}


func (r *NotificationRepository) SaveReminder(ctx context.Context, rem *domain.Reminder) error {
	query := `
		INSERT INTO reminders (event_id, user_id, message, scheduled_at, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	return r.pool.QueryRow(ctx, query,
		rem.EventID,
		rem.UserID,
		rem.Message,
		rem.ScheduledAt,
		rem.Status,
	).Scan(&rem.ID, &rem.CreatedAt)
}


func (r *NotificationRepository) GetPendingReminders(ctx context.Context, before time.Time) ([]*domain.Reminder, error) {
	query := `
		SELECT id, event_id, user_id, message, scheduled_at, status, created_at
		FROM reminders
		WHERE scheduled_at <= $1 AND status = 'pending'`

	rows, err := r.pool.Query(ctx, query, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []*domain.Reminder
	for rows.Next() {
		rem := &domain.Reminder{}
		if err := rows.Scan(
			&rem.ID,
			&rem.EventID,
			&rem.UserID,
			&rem.Message,
			&rem.ScheduledAt,
			&rem.Status,
			&rem.CreatedAt,
		); err != nil {
			return nil, err
		}
		reminders = append(reminders, rem)
	}
	return reminders, rows.Err()
}


func (r *NotificationRepository) UpdateReminderStatus(ctx context.Context, id, status string) error {
	query := `UPDATE reminders SET status = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}
