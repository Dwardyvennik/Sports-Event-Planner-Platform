package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/university/sports-event-planner-platform/services/notification-service/internal/domain"
	"github.com/university/sports-event-planner-platform/services/notification-service/internal/mailgun"
)


type NotificationRepository interface {
	Ping(context.Context) error
	SaveNotification(ctx context.Context, n *domain.Notification) error
	GetUserNotifications(ctx context.Context, userID string) ([]*domain.Notification, error)
	SaveReminder(ctx context.Context, rem *domain.Reminder) error
	GetPendingReminders(ctx context.Context, before time.Time) ([]*domain.Reminder, error)
	UpdateReminderStatus(ctx context.Context, id, status string) error
}


type SendNotificationInput struct {
	UserID  string
	Channel string
	Subject string
	Body    string
}


type SendReminderInput struct {
	EventID     string
	UserID      string
	Message     string
	ScheduledAt time.Time
}


type NotificationUseCase struct {
	notifications NotificationRepository
	mailgun       *mailgun.Client
	log           *slog.Logger
}


func NewNotificationUseCase(notifications NotificationRepository, mg *mailgun.Client, log *slog.Logger) *NotificationUseCase {
	return &NotificationUseCase{
		notifications: notifications,
		mailgun:       mg,
		log:           log,
	}
}


func (u *NotificationUseCase) Health(ctx context.Context) error {
	if u.notifications == nil {
		return nil
	}
	return u.notifications.Ping(ctx)
}


func (u *NotificationUseCase) SendNotification(ctx context.Context, input SendNotificationInput) error {
	if input.UserID == "" || input.Subject == "" {
		return domain.ErrInvalidNotification
	}

	n := &domain.Notification{
		UserID:  input.UserID,
		Channel: input.Channel,
		Subject: input.Subject,
		Body:    input.Body,
		Status:  domain.StatusPending,
	}

	var sendErr error

	switch input.Channel {
	case domain.ChannelEmail:
		if u.mailgun != nil && u.mailgun.IsConfigured() {
			sendErr = u.mailgun.Send(input.UserID, input.Subject, input.Body)
		} else {
			u.log.Warn("mailgun not configured, falling back to mock", "user_id", input.UserID)
			u.logMock(input.UserID, input.Subject, input.Body)
		}

	case domain.ChannelMock:
		u.logMock(input.UserID, input.Subject, input.Body)

	default:
		u.log.Warn("unknown notification channel, using mock", "channel", input.Channel)
		u.logMock(input.UserID, input.Subject, input.Body)
	}

	now := time.Now()
	if sendErr != nil {
		n.Status = domain.StatusFailed
		u.log.Error("notification send failed", "error", sendErr, "user_id", input.UserID, "channel", input.Channel)
	} else {
		n.Status = domain.StatusSent
		n.SentAt = &now
	}

	if err := u.notifications.SaveNotification(ctx, n); err != nil {
		return fmt.Errorf("save notification: %w", err)
	}
	return sendErr
}


func (u *NotificationUseCase) SendReminder(ctx context.Context, input SendReminderInput) error {
	if input.EventID == "" || input.UserID == "" {
		return domain.ErrInvalidNotification
	}

	rem := &domain.Reminder{
		EventID:     input.EventID,
		UserID:      input.UserID,
		Message:     input.Message,
		ScheduledAt: input.ScheduledAt,
		Status:      domain.StatusPending,
	}

	if err := u.notifications.SaveReminder(ctx, rem); err != nil {
		return fmt.Errorf("save reminder: %w", err)
	}

	
	if time.Until(input.ScheduledAt) <= time.Hour {
		u.log.Info("Reminder: "+input.Message+" for event "+input.EventID,
			"user_id", input.UserID,
			"event_id", input.EventID,
		)
		if err := u.notifications.UpdateReminderStatus(ctx, rem.ID, domain.StatusSent); err != nil {
			u.log.Error("update reminder status after immediate send", "error", err, "reminder_id", rem.ID)
		}
	}

	return nil
}


func (u *NotificationUseCase) GetNotifications(ctx context.Context, userID string) ([]*domain.Notification, error) {
	return u.notifications.GetUserNotifications(ctx, userID)
}


func (u *NotificationUseCase) ReminderWorker(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	u.log.Info("reminder worker started")

	for {
		select {
		case <-ctx.Done():
			u.log.Info("reminder worker stopped")
			return
		case t := <-ticker.C:
			u.processPendingReminders(ctx, t)
		}
	}
}

func (u *NotificationUseCase) processPendingReminders(ctx context.Context, now time.Time) {
	reminders, err := u.notifications.GetPendingReminders(ctx, now)
	if err != nil {
		u.log.Error("fetch pending reminders", "error", err)
		return
	}

	for _, rem := range reminders {
		u.log.Info("Reminder: "+rem.Message+" for event "+rem.EventID,
			"reminder_id", rem.ID,
			"user_id", rem.UserID,
			"scheduled_at", rem.ScheduledAt,
		)
		if err := u.notifications.UpdateReminderStatus(ctx, rem.ID, domain.StatusSent); err != nil {
			u.log.Error("update reminder status", "error", err, "reminder_id", rem.ID)
		}
	}
}


func (u *NotificationUseCase) logMock(userID, subject, body string) {
	u.log.Info("[MOCK NOTIFICATION]",
		"to", userID,
		"subject", subject,
		"body", body,
	)
}
