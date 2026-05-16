package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/university/sports-event-planner-platform/services/notification-service/internal/usecase"
)

const queueName = "event.created"


type eventCreatedMessage struct {
	EventID   string `json:"event_id"`
	Title     string `json:"title"`
	Sport     string `json:"sport"`
	StartTime string `json:"start_time"`
}


type EventConsumer struct {
	conn          *amqp.Connection
	notifications *usecase.NotificationUseCase
	log           *slog.Logger
}


func NewEventConsumer(conn *amqp.Connection, notifications *usecase.NotificationUseCase, log *slog.Logger) *EventConsumer {
	return &EventConsumer{
		conn:          conn,
		notifications: notifications,
		log:           log,
	}
}


func (c *EventConsumer) Start(ctx context.Context) {
	if c.conn == nil {
		c.log.Warn("rabbitmq not connected, event consumer disabled")
		return
	}

	ch, err := c.conn.Channel()
	if err != nil {
		c.log.Error("open rabbitmq channel", "error", err)
		return
	}
	defer ch.Close()

	
	if _, err := ch.QueueDeclare(
		queueName,
		true,  
		false, 
		false, 
		false, 
		nil,
	); err != nil {
		c.log.Error("declare queue", "queue", queueName, "error", err)
		return
	}

	msgs, err := ch.Consume(
		queueName,
		"notification-service", 
		false,                  
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		c.log.Error("start consuming", "queue", queueName, "error", err)
		return
	}

	c.log.Info("event consumer started", "queue", queueName)

	for {
		select {
		case <-ctx.Done():
			c.log.Info("event consumer stopped")
			return
		case msg, ok := <-msgs:
			if !ok {
				c.log.Warn("rabbitmq channel closed")
				return
			}
			c.handle(ctx, msg)
		}
	}
}

func (c *EventConsumer) handle(ctx context.Context, msg amqp.Delivery) {
	var payload eventCreatedMessage
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		c.log.Error("unmarshal event.created message", "error", err, "body", string(msg.Body))
		_ = msg.Nack(false, false) 
		return
	}

	input := usecase.SendNotificationInput{
		UserID:  payload.EventID, 
		Channel: "mock",
		Subject: fmt.Sprintf("New event: %s", payload.Title),
		Body:    fmt.Sprintf("A new %s event starts at %s", payload.Sport, payload.StartTime),
	}

	if err := c.notifications.SendNotification(ctx, input); err != nil {
		c.log.Error("send notification for event.created", "error", err, "event_id", payload.EventID)
		_ = msg.Nack(false, true) 
		return
	}

	_ = msg.Ack(false)
}
