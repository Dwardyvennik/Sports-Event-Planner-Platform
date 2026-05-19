package consumer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/university/sports-event-planner-platform/services/notification-service/internal/domain"
	"github.com/university/sports-event-planner-platform/services/notification-service/internal/usecase"
)

func TestEventConsumerHandleRetriesAndUsesPayloadUserID(t *testing.T) {
	repo := &retryNotificationRepository{failuresRemaining: 2}
	notifications := usecase.NewNotificationUseCase(repo, nil, discardLogger())
	consumer := NewEventConsumer(nil, notifications, discardLogger())

	consumer.handle(context.Background(), &nats.Msg{Data: []byte(`{
		"event_id":"event-123",
		"user_id":"user-456",
		"title":"Retry Match",
		"sport":"football",
		"start_time":"2026-06-01T10:00:00Z"
	}`)})

	if repo.saveAttempts != 3 {
		t.Fatalf("expected 3 save attempts after retries, got %d", repo.saveAttempts)
	}
	if len(repo.saved) != 3 {
		t.Fatalf("expected 3 saved notification attempts, got %d", len(repo.saved))
	}
	for _, notification := range repo.saved {
		if notification.UserID != "user-456" {
			t.Fatalf("expected payload user_id to be used, got %q", notification.UserID)
		}
		if notification.Subject != "New event: Retry Match" {
			t.Fatalf("unexpected notification subject: %q", notification.Subject)
		}
	}
}

func TestEventConsumerHandleInvalidJSONDoesNotSave(t *testing.T) {
	repo := &retryNotificationRepository{}
	notifications := usecase.NewNotificationUseCase(repo, nil, discardLogger())
	consumer := NewEventConsumer(nil, notifications, discardLogger())

	consumer.handle(context.Background(), &nats.Msg{Data: []byte(`not-json`)})

	if repo.saveAttempts != 0 {
		t.Fatalf("expected no save attempts for invalid json, got %d", repo.saveAttempts)
	}
}

type retryNotificationRepository struct {
	mu                sync.Mutex
	failuresRemaining int
	saveAttempts      int
	saved             []*domain.Notification
}

func (r *retryNotificationRepository) Ping(context.Context) error {
	return nil
}

func (r *retryNotificationRepository) SaveNotification(_ context.Context, notification *domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.saveAttempts++
	copied := *notification
	r.saved = append(r.saved, &copied)

	if r.failuresRemaining > 0 {
		r.failuresRemaining--
		return errors.New("temporary save failure")
	}
	return nil
}

func (r *retryNotificationRepository) GetUserNotifications(context.Context, string) ([]*domain.Notification, error) {
	return nil, nil
}

func (r *retryNotificationRepository) SaveReminder(context.Context, *domain.Reminder) error {
	return nil
}

func (r *retryNotificationRepository) GetPendingReminders(context.Context, time.Time) ([]*domain.Reminder, error) {
	return nil, nil
}

func (r *retryNotificationRepository) UpdateReminderStatus(context.Context, string, string) error {
	return nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
