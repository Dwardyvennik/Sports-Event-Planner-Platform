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

	"github.com/dwardyvennik/sports-event-planner-platform/services/notification-service/internal/domain"
	"github.com/dwardyvennik/sports-event-planner-platform/services/notification-service/internal/usecase"
)

func TestEventConsumerHandleRetriesAndUsesPayloadUserID(t *testing.T) {
	repo := &retryNotificationRepository{failuresRemaining: 2}
	notifications := usecase.NewNotificationUseCase(repo, nil, nil, discardLogger())
	consumer := NewEventConsumer(nil, notifications, discardLogger(), "")

	consumer.handle(context.Background(), eventSubscription{subject: eventCreatedSubject, action: "created"}, &nats.Msg{Data: []byte(`{
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
	notifications := usecase.NewNotificationUseCase(repo, nil, nil, discardLogger())
	consumer := NewEventConsumer(nil, notifications, discardLogger(), "")

	consumer.handle(context.Background(), eventSubscription{subject: eventCreatedSubject, action: "created"}, &nats.Msg{Data: []byte(`not-json`)})

	if repo.saveAttempts != 0 {
		t.Fatalf("expected no save attempts for invalid json, got %d", repo.saveAttempts)
	}
}

func TestEventConsumerHandleUpdatedAndJoinedSubjects(t *testing.T) {
	for _, tc := range []struct {
		name    string
		sub     eventSubscription
		subject string
	}{
		{name: "updated", sub: eventSubscription{subject: eventUpdatedSubject, action: "updated"}, subject: "Event updated: Demo Match"},
		{name: "joined", sub: eventSubscription{subject: eventJoinedSubject, action: "joined"}, subject: "Joined event: Demo Match"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &retryNotificationRepository{}
			notifications := usecase.NewNotificationUseCase(repo, nil, nil, discardLogger())
			consumer := NewEventConsumer(nil, notifications, discardLogger(), "")

			consumer.handle(context.Background(), tc.sub, &nats.Msg{Data: []byte(`{
				"event_id":"event-123",
				"user_id":"user-456",
				"title":"Demo Match",
				"sport":"football",
				"start_time":"2026-06-01T10:00:00Z"
			}`)})

			if repo.saveAttempts != 1 {
				t.Fatalf("expected one notification save, got %d", repo.saveAttempts)
			}
			if repo.saved[0].Subject != tc.subject {
				t.Fatalf("expected subject %q, got %q", tc.subject, repo.saved[0].Subject)
			}
		})
	}
}

func TestEventConsumer_UsesConfiguredChannel(t *testing.T) {
	repo := &retryNotificationRepository{}
	notifications := usecase.NewNotificationUseCase(repo, nil, configuredEmailSender{}, discardLogger())
	consumer := NewEventConsumer(nil, notifications, discardLogger(), domain.ChannelEmail)

	consumer.handle(context.Background(), eventSubscription{subject: eventCreatedSubject, action: "created"}, &nats.Msg{Data: []byte(`{
		"event_id":"event-123",
		"user_id":"user-456",
		"title":"Email Match",
		"sport":"football",
		"start_time":"2026-06-01T10:00:00Z"
	}`)})

	if repo.saveAttempts != 1 {
		t.Fatalf("expected one notification save, got %d", repo.saveAttempts)
	}
	if repo.saved[0].Channel != domain.ChannelEmail {
		t.Fatalf("expected configured channel 'email', got %q", repo.saved[0].Channel)
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

type configuredEmailSender struct{}

func (configuredEmailSender) Send(string, string, string) error {
	return nil
}

func (configuredEmailSender) IsConfigured() bool {
	return true
}
