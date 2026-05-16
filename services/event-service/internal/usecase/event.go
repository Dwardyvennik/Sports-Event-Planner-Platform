package usecase

<<<<<<< Updated upstream
import "context"

type EventRepository interface {
	Ping(context.Context) error
=======
import (
	"context"
	"strings"
	"time"

	"github.com/university/sports-event-planner-platform/services/event-service/internal/domain"
)

type EventRepository interface {
	Ping(context.Context) error
	Create(context.Context, *domain.Event) error
	Get(context.Context, string) (*domain.Event, error)
	List(context.Context, domain.EventFilter) ([]*domain.Event, error)
	Delete(context.Context, string) error
}

type EventService interface {
	Health(context.Context) error
	CreateEvent(context.Context, CreateEventInput) (*domain.Event, error)
	GetEvent(context.Context, string) (*domain.Event, error)
	ListEvents(context.Context, ListEventsInput) ([]*domain.Event, error)
	DeleteEvent(context.Context, string) error
}

type CreateEventInput struct {
	Sport       string
	Category    string
	Competition string
	Title       string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Status      string
	Country     string
	City        string
}

type ListEventsInput struct {
	Sport         string
	Competition   string
	StartTimeFrom time.Time
	StartTimeTo   time.Time
	Country       string
	Page          int
	PageSize      int
>>>>>>> Stashed changes
}

type EventUseCase struct {
	events EventRepository
}

func NewEventUseCase(events EventRepository) *EventUseCase {
	return &EventUseCase{events: events}
}

func (u *EventUseCase) Health(ctx context.Context) error {
	if u.events == nil {
		return nil
	}
	return u.events.Ping(ctx)
}
<<<<<<< Updated upstream
=======

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
	return event, nil
}

func (u *EventUseCase) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, domain.ErrEventNotFound
	}
	return u.events.Get(ctx, id)
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

func (u *EventUseCase) DeleteEvent(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.ErrEventNotFound
	}
	return u.events.Delete(ctx, id)
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
>>>>>>> Stashed changes
