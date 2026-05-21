package tests

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/dwardyvennik/sports-event-planner-platform/services/notification-service/internal/domain"
	"github.com/dwardyvennik/sports-event-planner-platform/services/notification-service/internal/usecase"
)

type mockRepo struct {
	notifications []*domain.Notification
	reminders     []*domain.Reminder
	pingErr       error
	saveErr       error
}

func (m *mockRepo) Ping(_ context.Context) error { return m.pingErr }

func (m *mockRepo) SaveNotification(_ context.Context, n *domain.Notification) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	n.ID = "test-id-123"
	n.CreatedAt = time.Now()
	m.notifications = append(m.notifications, n)
	return nil
}

func (m *mockRepo) GetUserNotifications(_ context.Context, userID string) ([]*domain.Notification, error) {
	var result []*domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID {
			result = append(result, n)
		}
	}
	return result, nil
}

func (m *mockRepo) SaveReminder(_ context.Context, rem *domain.Reminder) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	rem.ID = "reminder-id-123"
	rem.CreatedAt = time.Now()
	m.reminders = append(m.reminders, rem)
	return nil
}

func (m *mockRepo) GetPendingReminders(_ context.Context, before time.Time) ([]*domain.Reminder, error) {
	var result []*domain.Reminder
	for _, r := range m.reminders {
		if r.Status == domain.StatusPending && !r.ScheduledAt.After(before) {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockRepo) UpdateReminderStatus(_ context.Context, id, status string) error {
	for _, r := range m.reminders {
		if r.ID == id {
			r.Status = status
		}
	}
	return nil
}

func newTestUseCase(repo *mockRepo) *usecase.NotificationUseCase {
	return usecase.NewNotificationUseCase(repo, nil, nil, discardLogger())
}

type stubEmailSender struct {
	configured bool
	sends      int
	to         string
	subject    string
	body       string
	err        error
}

func (s *stubEmailSender) Send(to, subject, body string) error {
	s.sends++
	s.to = to
	s.subject = subject
	s.body = body
	return s.err
}

func (s *stubEmailSender) IsConfigured() bool {
	return s.configured
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// --- Tests ---

func TestHealth_OK(t *testing.T) {
	repo := &mockRepo{}
	uc := newTestUseCase(repo)
	if err := uc.Health(context.Background()); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestHealth_Error(t *testing.T) {
	repo := &mockRepo{pingErr: errors.New("db down")}
	uc := newTestUseCase(repo)
	if err := uc.Health(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSendNotification_MockChannel(t *testing.T) {
	repo := &mockRepo{}
	uc := newTestUseCase(repo)

	_, err := uc.SendNotification(context.Background(), usecase.SendNotificationInput{
		UserID:  "user-1",
		Channel: domain.ChannelMock,
		Subject: "Test Subject",
		Body:    "Test Body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.notifications) != 1 {
		t.Fatalf("expected 1 notification saved, got %d", len(repo.notifications))
	}
	if repo.notifications[0].Status != domain.StatusSent {
		t.Fatalf("expected status 'sent', got '%s'", repo.notifications[0].Status)
	}
}

func TestSendNotification_FallsBackToSMTP(t *testing.T) {
	repo := &mockRepo{}
	mg := &stubEmailSender{configured: false}
	smtp := &stubEmailSender{configured: true}
	uc := usecase.NewNotificationUseCase(repo, mg, smtp, discardLogger())

	n, err := uc.SendNotification(context.Background(), usecase.SendNotificationInput{
		UserID:  "user-1",
		Channel: domain.ChannelEmail,
		Subject: "Test Subject",
		Body:    "<p>Test Body</p>",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Status != domain.StatusSent {
		t.Fatalf("expected status 'sent', got '%s'", n.Status)
	}
	if n.Channel != domain.ChannelEmail {
		t.Fatalf("expected channel 'email', got '%s'", n.Channel)
	}
	if mg.sends != 0 {
		t.Fatalf("expected mailgun not to send, got %d sends", mg.sends)
	}
	if smtp.sends != 1 {
		t.Fatalf("expected smtp to send once, got %d sends", smtp.sends)
	}
	if smtp.to != "user-1@dev.local" {
		t.Fatalf("expected dev email to be resolved, got %q", smtp.to)
	}
}

func TestSendNotification_FallsBackToMock(t *testing.T) {
	repo := &mockRepo{}
	mg := &stubEmailSender{configured: false}
	smtp := &stubEmailSender{configured: false}
	uc := usecase.NewNotificationUseCase(repo, mg, smtp, discardLogger())

	n, err := uc.SendNotification(context.Background(), usecase.SendNotificationInput{
		UserID:  "user-1",
		Channel: domain.ChannelEmail,
		Subject: "Test Subject",
		Body:    "Test Body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Status != domain.StatusSent {
		t.Fatalf("expected status 'sent', got '%s'", n.Status)
	}
	if n.Channel != domain.ChannelMock {
		t.Fatalf("expected channel 'mock', got '%s'", n.Channel)
	}
	if len(repo.notifications) != 1 {
		t.Fatalf("expected 1 notification saved, got %d", len(repo.notifications))
	}
}

func TestSendNotification_MissingUserID(t *testing.T) {
	repo := &mockRepo{}
	uc := newTestUseCase(repo)

	_, err := uc.SendNotification(context.Background(), usecase.SendNotificationInput{
		UserID:  "",
		Channel: domain.ChannelMock,
		Subject: "Test",
		Body:    "Body",
	})
	if !errors.Is(err, domain.ErrInvalidNotification) {
		t.Fatalf("expected ErrInvalidNotification, got: %v", err)
	}
}

func TestSendNotification_MissingSubject(t *testing.T) {
	repo := &mockRepo{}
	uc := newTestUseCase(repo)

	_, err := uc.SendNotification(context.Background(), usecase.SendNotificationInput{
		UserID:  "user-1",
		Channel: domain.ChannelMock,
		Subject: "",
		Body:    "Body",
	})
	if !errors.Is(err, domain.ErrInvalidNotification) {
		t.Fatalf("expected ErrInvalidNotification, got: %v", err)
	}
}

func TestGetNotifications(t *testing.T) {
	repo := &mockRepo{}
	uc := newTestUseCase(repo)

	// Send two notifications for user-1, one for user-2
	_, _ = uc.SendNotification(context.Background(), usecase.SendNotificationInput{
		UserID: "user-1", Channel: domain.ChannelMock, Subject: "S1", Body: "B1",
	})
	_, _ = uc.SendNotification(context.Background(), usecase.SendNotificationInput{
		UserID: "user-1", Channel: domain.ChannelMock, Subject: "S2", Body: "B2",
	})
	_, _ = uc.SendNotification(context.Background(), usecase.SendNotificationInput{
		UserID: "user-2", Channel: domain.ChannelMock, Subject: "S3", Body: "B3",
	})

	result, err := uc.GetNotifications(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 notifications for user-1, got %d", len(result))
	}
}

func TestSendReminder_Immediate(t *testing.T) {
	repo := &mockRepo{}
	uc := newTestUseCase(repo)

	// ScheduledAt within 1 hour → should be sent immediately
	err := uc.SendReminder(context.Background(), usecase.SendReminderInput{
		EventID:     "event-1",
		UserID:      "user-1",
		Message:     "Your event is soon!",
		ScheduledAt: time.Now().Add(30 * time.Minute),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.reminders) != 1 {
		t.Fatalf("expected 1 reminder saved, got %d", len(repo.reminders))
	}
	// Status should be updated to "sent" since it's within 1 hour
	if repo.reminders[0].Status != domain.StatusSent {
		t.Fatalf("expected status 'sent', got '%s'", repo.reminders[0].Status)
	}
}

func TestSendReminder_Scheduled(t *testing.T) {
	repo := &mockRepo{}
	uc := newTestUseCase(repo)

	// ScheduledAt more than 1 hour away → should stay pending
	err := uc.SendReminder(context.Background(), usecase.SendReminderInput{
		EventID:     "event-2",
		UserID:      "user-1",
		Message:     "Your event is tomorrow!",
		ScheduledAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.reminders[0].Status != domain.StatusPending {
		t.Fatalf("expected status 'pending', got '%s'", repo.reminders[0].Status)
	}
}

func TestSendReminder_MissingFields(t *testing.T) {
	repo := &mockRepo{}
	uc := newTestUseCase(repo)

	err := uc.SendReminder(context.Background(), usecase.SendReminderInput{
		EventID: "",
		UserID:  "",
	})
	if !errors.Is(err, domain.ErrInvalidNotification) {
		t.Fatalf("expected ErrInvalidNotification, got: %v", err)
	}
}

func TestDomainConstants(t *testing.T) {
	if domain.ChannelEmail != "email" {
		t.Error("ChannelEmail should be 'email'")
	}
	if domain.ChannelMock != "mock" {
		t.Error("ChannelMock should be 'mock'")
	}
	if domain.StatusPending != "pending" {
		t.Error("StatusPending should be 'pending'")
	}
	if domain.StatusSent != "sent" {
		t.Error("StatusSent should be 'sent'")
	}
	if domain.StatusFailed != "failed" {
		t.Error("StatusFailed should be 'failed'")
	}
}
