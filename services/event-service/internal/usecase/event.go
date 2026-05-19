package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"github.com/university/sports-event-planner-platform/services/event-service/internal/domain"
)

const (
	subjectEventCreated = "events.created"
	subjectEventUpdated = "events.updated"
	subjectEventJoined  = "events.joined"
)

type EventRepository interface {
	Ping(context.Context) error
	Create(context.Context, *domain.Event) error
	Get(context.Context, string) (*domain.Event, error)
	List(context.Context, domain.EventFilter) ([]*domain.Event, error)
	Update(context.Context, *domain.Event) error
	Delete(context.Context, string, string) error
	Join(context.Context, string, string) (*domain.Event, error)
	Leave(context.Context, string, string) (*domain.Event, error)
}

type EventService interface {
	Health(context.Context) error
	CreateEvent(context.Context, CreateEventInput) (*domain.Event, error)
	GetEvent(context.Context, string) (*domain.Event, error)
	ListEvents(context.Context, ListEventsInput) ([]*domain.Event, error)
	UpdateEvent(context.Context, string, CreateEventInput) (*domain.Event, error)
	DeleteEvent(context.Context, string, string) error
	JoinEvent(context.Context, string, string) (*domain.Event, error)
	LeaveEvent(context.Context, string, string) (*domain.Event, error)
}

type CreateEventInput struct {
	CreatorID       string
	Sport           string
	Category        string
	Competition     string
	Title           string
	Description     string
	Location        string
	StartTime       time.Time
	EndTime         time.Time
	Status          string
	Country         string
	City            string
	MaxParticipants int32
}

type ListEventsInput struct {
	Sport         string
	Category      string
	Competition   string
	Status        string
	StartTimeFrom time.Time
	StartTimeTo   time.Time
	Country       string
	City          string
	Page          int
	PageSize      int
}

type EventUseCase struct {
	events EventRepository
	cache  *redis.Client
	broker *nats.Conn
	log    *slog.Logger
}

func NewEventUseCase(events EventRepository, cache *redis.Client, broker *nats.Conn, log *slog.Logger) *EventUseCase {
	if log == nil {
		log = slog.Default()
	}
	return &EventUseCase{
		events: events,
		cache:  cache,
		broker: broker,
		log:    log,
	}
}

func (u *EventUseCase) Health(ctx context.Context) error {
	if u.events == nil {
		return nil
	}
	return u.events.Ping(ctx)
}

func (u *EventUseCase) CreateEvent(ctx context.Context, input CreateEventInput) (*domain.Event, error) {
	event := eventFromInput("", input.CreatorID, input)
	if err := validateEvent(event); err != nil {
		return nil, err
	}
	if err := u.events.Create(ctx, event); err != nil {
		return nil, err
	}
	u.publish(ctx, subjectEventCreated, eventPayload(event))
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
		Category:      normalize(input.Category),
		Competition:   normalize(input.Competition),
		Status:        normalize(input.Status),
		StartTimeFrom: input.StartTimeFrom,
		StartTimeTo:   input.StartTimeTo,
		Country:       normalize(input.Country),
		City:          strings.TrimSpace(input.City),
		Limit:         pageSize,
		Offset:        (page - 1) * pageSize,
	})
}

func (u *EventUseCase) UpdateEvent(ctx context.Context, id string, input CreateEventInput) (*domain.Event, error) {
	event := eventFromInput(id, input.CreatorID, input)
	if err := validateEvent(event); err != nil {
		return nil, err
	}
	if err := u.events.Update(ctx, event); err != nil {
		return nil, err
	}
	u.deleteEventCache(ctx, event.ID)

	updated, err := u.events.Get(ctx, event.ID)
	if err != nil {
		return nil, err
	}
	u.publish(ctx, subjectEventUpdated, eventPayload(updated))
	return updated, nil
}

func (u *EventUseCase) DeleteEvent(ctx context.Context, id string, userID string) error {
	id = strings.TrimSpace(id)
	userID = strings.TrimSpace(userID)
	if id == "" {
		return domain.ErrEventNotFound
	}
	if userID == "" {
		return domain.ErrEventForbidden
	}
	if err := u.events.Delete(ctx, id, userID); err != nil {
		return err
	}
	u.deleteEventCache(ctx, id)
	return nil
}

func (u *EventUseCase) JoinEvent(ctx context.Context, id string, userID string) (*domain.Event, error) {
	id = strings.TrimSpace(id)
	userID = strings.TrimSpace(userID)
	if id == "" {
		return nil, domain.ErrEventNotFound
	}
	if userID == "" {
		return nil, domain.ErrInvalidEvent
	}
	event, err := u.events.Join(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	u.deleteEventCache(ctx, id)
	u.publish(ctx, subjectEventJoined, map[string]string{
		"event_id": id,
		"user_id":  userID,
	})
	return event, nil
}

func (u *EventUseCase) LeaveEvent(ctx context.Context, id string, userID string) (*domain.Event, error) {
	id = strings.TrimSpace(id)
	userID = strings.TrimSpace(userID)
	if id == "" {
		return nil, domain.ErrEventNotFound
	}
	if userID == "" {
		return nil, domain.ErrInvalidEvent
	}
	event, err := u.events.Leave(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	u.deleteEventCache(ctx, id)
	return event, nil
}

func eventFromInput(id string, creatorID string, input CreateEventInput) *domain.Event {
	return &domain.Event{
		ID:              strings.TrimSpace(id),
		CreatorID:       strings.TrimSpace(creatorID),
		Sport:           normalize(input.Sport),
		Category:        normalize(input.Category),
		Competition:     normalize(input.Competition),
		Title:           strings.TrimSpace(input.Title),
		Description:     strings.TrimSpace(input.Description),
		Location:        strings.TrimSpace(input.Location),
		StartTime:       input.StartTime,
		EndTime:         input.EndTime,
		Status:          normalizeStatus(input.Status),
		Country:         normalize(input.Country),
		City:            strings.TrimSpace(input.City),
		MaxParticipants: input.MaxParticipants,
	}
}

func validateEvent(event *domain.Event) error {
	if event.CreatorID == "" || event.Title == "" || event.Sport == "" || event.StartTime.IsZero() {
		return domain.ErrInvalidEvent
	}
	if !event.EndTime.IsZero() && event.EndTime.Before(event.StartTime) {
		return domain.ErrInvalidEvent
	}
	if event.MaxParticipants < 0 {
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

func (u *EventUseCase) publish(ctx context.Context, subject string, payload any) {
	if u.broker == nil {
		return
	}
	body, err := json.Marshal(payload)
	if err != nil {
		u.log.WarnContext(ctx, "marshal event payload", "subject", subject, "error", err)
		return
	}
	for attempt := 1; attempt <= 3; attempt++ {
		if err := u.broker.Publish(subject, body); err != nil {
			u.log.WarnContext(ctx, "publish nats event failed", "subject", subject, "attempt", attempt, "error", err)
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
			continue
		}
		if err := u.broker.FlushWithContext(ctx); err != nil {
			u.log.WarnContext(ctx, "flush nats event failed", "subject", subject, "attempt", attempt, "error", err)
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
			continue
		}
		return
	}
}

func eventPayload(event *domain.Event) map[string]any {
	return map[string]any{
		"event_id":           event.ID,
		"user_id":            event.CreatorID,
		"title":              event.Title,
		"sport":              event.Sport,
		"start_time":         event.StartTime.UTC().Format(time.RFC3339),
		"participants_count": event.ParticipantsCount,
	}
}
