package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"

	"github.com/university/sports-event-planner-platform/services/event-service/internal/domain"
)

type EventRepository interface {
	Ping(context.Context) error
	Create(context.Context, *domain.Event) error
	Get(context.Context, string) (*domain.Event, error)
	List(context.Context, domain.EventFilter) ([]*domain.Event, error)
	UpdateEvent(context.Context, string, string, string, string, time.Time, time.Time, int32) (*domain.Event, error)
	Delete(context.Context, string) error
}

type EventService interface {
	Health(context.Context) error
	CreateEvent(context.Context, CreateEventInput) (*domain.Event, error)
	GetEvent(context.Context, string) (*domain.Event, error)
	ListEvents(context.Context, ListEventsInput) ([]*domain.Event, error)
	UpdateEvent(context.Context, string, CreateEventInput) (*domain.Event, error)
	DeleteEvent(context.Context, string) error
}

type CreateEventInput struct {
	Sport           string
	Category        string
	Competition     string
	Title           string
	Description     string
	StartTime       time.Time
	EndTime         time.Time
	Status          string
	Country         string
	City            string
	Location        string
	MaxParticipants int32
	CreatorID       string
}

type ListEventsInput struct {
	Sport         string
	Competition   string
	StartTimeFrom time.Time
	StartTimeTo   time.Time
	Country       string
	Page          int
	PageSize      int
}

type EventUseCase struct {
	events EventRepository
	cache  *redis.Client
	broker *amqp.Connection
}

func NewEventUseCase(events EventRepository, cache *redis.Client, broker *amqp.Connection) *EventUseCase {
	return &EventUseCase{
		events: events,
		cache:  cache,
		broker: broker,
	}
}

func (u *EventUseCase) Health(ctx context.Context) error {
	if u.events == nil {
		return nil
	}
	return u.events.Ping(ctx)
}

func (u *EventUseCase) CreateEvent(ctx context.Context, input CreateEventInput) (*domain.Event, error) {
	event := &domain.Event{
		Sport:       normalize(input.Sport),
		Category:    normalize(input.Category),
		Competition: normalize(input.Competition),
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
		Status:      normalizeStatus(input.Status),
		Country:     normalize(input.Country),
		City:        strings.TrimSpace(input.City),
	}
	if err := validateEvent(event); err != nil {
		return nil, err
	}
	if err := u.events.Create(ctx, event); err != nil {
		return nil, err
	}
	payload := eventCreatedPayload{
		EventID:   event.ID,
		UserID:    strings.TrimSpace(input.CreatorID),
		Title:     event.Title,
		Sport:     event.Sport,
		StartTime: event.StartTime.UTC().Format(time.RFC3339),
	}
	if err := publishEvent(ctx, u.broker, payload); err != nil {
		slog.Default().Warn("publish event.created failed", "error", err)
	}
	return event, nil
}

func (u *EventUseCase) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, domain.ErrEventNotFound
	}
	if event, ok := u.eventFromCache(ctx, id); ok {
		return event, nil
	}
	event, err := u.events.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	u.cacheEvent(ctx, event)
	return event, nil
}

func (u *EventUseCase) ListEvents(ctx context.Context, input ListEventsInput) ([]*domain.Event, error) {
	if !input.StartTimeFrom.IsZero() && !input.StartTimeTo.IsZero() && input.StartTimeTo.Before(input.StartTimeFrom) {
		return nil, domain.ErrInvalidEvent
	}

	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return u.events.List(ctx, domain.EventFilter{
		Sport:         normalize(input.Sport),
		Competition:   normalize(input.Competition),
		StartTimeFrom: input.StartTimeFrom,
		StartTimeTo:   input.StartTimeTo,
		Country:       normalize(input.Country),
		Limit:         pageSize,
		Offset:        (page - 1) * pageSize,
	})
}

func (u *EventUseCase) UpdateEvent(ctx context.Context, id string, input CreateEventInput) (*domain.Event, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, domain.ErrEventNotFound
	}
	startTime := input.StartTime
	endTime := input.EndTime
	if !endTime.IsZero() && endTime.Before(startTime) {
		return nil, domain.ErrInvalidEvent
	}
	title := strings.TrimSpace(input.Title)
	sport := normalize(input.Sport)
	if title == "" || sport == "" || startTime.IsZero() {
		return nil, domain.ErrInvalidEvent
	}

	event, err := u.events.UpdateEvent(
		ctx,
		id,
		title,
		sport,
		strings.TrimSpace(input.Location),
		startTime,
		endTime,
		input.MaxParticipants,
	)
	if err != nil {
		return nil, err
	}
	u.deleteEventCache(ctx, id)
	return event, nil
}

func (u *EventUseCase) DeleteEvent(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.ErrEventNotFound
	}
	if err := u.events.Delete(ctx, id); err != nil {
		return err
	}
	u.deleteEventCache(ctx, id)
	return nil
}

func validateEvent(event *domain.Event) error {
	if event.Title == "" || event.Sport == "" || event.StartTime.IsZero() {
		return domain.ErrInvalidEvent
	}
	if !event.EndTime.IsZero() && event.EndTime.Before(event.StartTime) {
		return domain.ErrInvalidEvent
	}
	return nil
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeStatus(status string) string {
	status = normalize(status)
	if status == "" {
		return "scheduled"
	}
	return status
}

func (u *EventUseCase) eventFromCache(ctx context.Context, id string) (*domain.Event, bool) {
	if u.cache == nil {
		return nil, false
	}
	data, err := u.cache.Get(ctx, "event:"+id).Bytes()
	if err != nil {
		return nil, false
	}
	event := new(domain.Event)
	if err := json.Unmarshal(data, event); err != nil {
		return nil, false
	}
	return event, true
}

func (u *EventUseCase) cacheEvent(ctx context.Context, event *domain.Event) {
	if u.cache == nil || event == nil || event.ID == "" {
		return
	}
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	_ = u.cache.Set(ctx, "event:"+event.ID, data, 5*time.Minute).Err()
}

func (u *EventUseCase) deleteEventCache(ctx context.Context, id string) {
	if u.cache == nil {
		return
	}
	_ = u.cache.Del(ctx, "event:"+id).Err()
}

type eventCreatedPayload struct {
	EventID   string `json:"event_id"`
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	Sport     string `json:"sport"`
	StartTime string `json:"start_time"`
}

func publishEvent(ctx context.Context, conn *amqp.Connection, payload eventCreatedPayload) error {
	if conn == nil {
		return nil
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if _, err := ch.QueueDeclare(
		"event.created",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	return ch.PublishWithContext(
		ctx,
		"",
		"event.created",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
