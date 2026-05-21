package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/dwardyvennik/sports-event-planner-platform/pkg/metrics"
	"github.com/dwardyvennik/sports-event-planner-platform/services/notification-service/internal/usecase"
)

const (
	eventCreatedSubject = "events.created"
	eventUpdatedSubject = "events.updated"
	eventJoinedSubject  = "events.joined"
	defaultChannel      = "mock"
)

type eventMessage struct {
	EventID   string `json:"event_id"`
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	Sport     string `json:"sport"`
	StartTime string `json:"start_time"`
}

type eventSubscription struct {
	subject string
	action  string
}

type EventConsumer struct {
	conn                *nats.Conn
	notifications       *usecase.NotificationUseCase
	log                 *slog.Logger
	notificationChannel string
}

func NewEventConsumer(conn *nats.Conn, notifications *usecase.NotificationUseCase, log *slog.Logger, channel string) *EventConsumer {
	if channel == "" {
		channel = defaultChannel
	}
	return &EventConsumer{
		conn:                conn,
		notifications:       notifications,
		log:                 log,
		notificationChannel: channel,
	}
}

func (c *EventConsumer) Start(ctx context.Context) {
	if c.conn == nil {
		c.log.Warn("nats not connected, event consumer disabled")
		return
	}

	subscriptions := []eventSubscription{
		{subject: eventCreatedSubject, action: "created"},
		{subject: eventUpdatedSubject, action: "updated"},
		{subject: eventJoinedSubject, action: "joined"},
	}
	subs := make([]*nats.Subscription, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		subscription := subscription
		sub, err := c.conn.Subscribe(subscription.subject, func(msg *nats.Msg) {
			c.handle(ctx, subscription, msg)
		})
		if err != nil {
			c.log.Error("subscribe to nats subject", "subject", subscription.subject, "error", err)
			return
		}
		subs = append(subs, sub)
		c.log.Info("event consumer started", "subject", subscription.subject)
	}
	defer func() {
		for _, sub := range subs {
			_ = sub.Unsubscribe()
		}
	}()

	<-ctx.Done()

	for _, sub := range subs {
		if err := sub.Drain(); err != nil {
			c.log.Warn("drain nats subscription", "subject", sub.Subject, "error", err)
		}
	}
	c.log.Info("event consumer stopped")
}

func (c *EventConsumer) handle(ctx context.Context, subscription eventSubscription, msg *nats.Msg) {
	var payload eventMessage
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		metrics.NATSConsumerFailedTotal.WithLabelValues(subscription.subject).Inc()
		c.log.Error("unmarshal nats event message", "subject", subscription.subject, "error", err, "body", string(msg.Data))
		return
	}

	input := usecase.SendNotificationInput{
		UserID:  payload.UserID,
		Channel: c.notificationChannel,
		Subject: notificationSubject(subscription.action, payload),
		Body:    notificationBody(subscription.action, payload),
	}

	for attempt := 1; attempt <= 3; attempt++ {
		if _, err := c.notifications.SendNotification(ctx, input); err != nil {
			metrics.NATSConsumerRetryTotal.WithLabelValues(subscription.subject).Inc()
			c.log.Warn("send notification for nats event failed", "subject", subscription.subject, "error", err, "event_id", payload.EventID, "attempt", attempt)
			time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
			continue
		}
		metrics.NATSConsumedTotal.WithLabelValues(subscription.subject).Inc()
		c.log.Info("nats event consumed", "subject", subscription.subject, "event_id", payload.EventID, "user_id", payload.UserID)
		return
	}

	metrics.NATSConsumerFailedTotal.WithLabelValues(subscription.subject).Inc()
	c.log.Error("send notification for nats event exhausted retries", "subject", subscription.subject, "event_id", payload.EventID, "user_id", payload.UserID)
}

func notificationSubject(action string, payload eventMessage) string {
	switch action {
	case "created":
		return fmt.Sprintf("New event: %s", payload.Title)
	case "updated":
		return fmt.Sprintf("Event updated: %s", payload.Title)
	case "joined":
		return fmt.Sprintf("Joined event: %s", payload.Title)
	default:
		return fmt.Sprintf("Event notification: %s", payload.Title)
	}
}

func notificationBody(action string, payload eventMessage) string {
	switch action {
	case "created":
		return fmt.Sprintf("A new %s event starts at %s", payload.Sport, payload.StartTime)
	case "updated":
		return fmt.Sprintf("The %s event was updated. Start time: %s", payload.Sport, payload.StartTime)
	case "joined":
		return fmt.Sprintf("You joined %s. Start time: %s", payload.Title, payload.StartTime)
	default:
		return fmt.Sprintf("Event %s starts at %s", payload.Title, payload.StartTime)
	}
}
