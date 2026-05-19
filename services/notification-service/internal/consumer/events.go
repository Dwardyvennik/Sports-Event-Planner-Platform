package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/university/sports-event-planner-platform/services/notification-service/internal/domain"
	"github.com/university/sports-event-planner-platform/services/notification-service/internal/usecase"
)

const eventCreatedSubject = "events.created"

type eventCreatedMessage struct {
	EventID   string `json:"event_id"`
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	Sport     string `json:"sport"`
	StartTime string `json:"start_time"`
}

type EventConsumer struct {
	conn          *nats.Conn
	notifications *usecase.NotificationUseCase
	log           *slog.Logger
}

func NewEventConsumer(conn *nats.Conn, notifications *usecase.NotificationUseCase, log *slog.Logger) *EventConsumer {
	return &EventConsumer{
		conn:          conn,
		notifications: notifications,
		log:           log,
	}
}

func (c *EventConsumer) Start(ctx context.Context) {
	if c.conn == nil {
		c.log.Warn("nats not connected, event consumer disabled")
		return
	}

	sub, err := c.conn.Subscribe(eventCreatedSubject, func(msg *nats.Msg) {
		c.handle(ctx, msg)
	})
	if err != nil {
		c.log.Error("subscribe to nats subject", "subject", eventCreatedSubject, "error", err)
		return
	}
	defer sub.Unsubscribe()

	c.log.Info("event consumer started", "subject", eventCreatedSubject)
	<-ctx.Done()

	if err := sub.Drain(); err != nil {
		c.log.Warn("drain nats subscription", "subject", eventCreatedSubject, "error", err)
	}
	c.log.Info("event consumer stopped")
}

func (c *EventConsumer) handle(ctx context.Context, msg *nats.Msg) {
	var payload eventCreatedMessage
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		c.log.Error("unmarshal events.created message", "error", err, "body", string(msg.Data))
		return
	}

	input := usecase.SendNotificationInput{
		UserID:  payload.UserID,
		Channel: domain.ChannelMock,
		Subject: fmt.Sprintf("New event: %s", payload.Title),
		Body:    fmt.Sprintf("A new %s event starts at %s", payload.Sport, payload.StartTime),
	}

	for attempt := 1; attempt <= 3; attempt++ {
		if err := c.notifications.SendNotification(ctx, input); err != nil {
			c.log.Warn("send notification for events.created failed", "error", err, "event_id", payload.EventID, "attempt", attempt)
			time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
			continue
		}
		return
	}

	c.log.Error("send notification for events.created exhausted retries", "event_id", payload.EventID, "user_id", payload.UserID)
}
