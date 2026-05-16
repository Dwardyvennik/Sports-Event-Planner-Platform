package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/university/sports-event-planner-platform/services/event-service/internal/domain"
)

type EventRepository struct {
	mu     sync.RWMutex
	events map[string]*domain.Event
}

func NewEventRepository() *EventRepository {
	return &EventRepository{events: make(map[string]*domain.Event)}
}

func (r *EventRepository) Ping(context.Context) error {
	return nil
}

func (r *EventRepository) Create(_ context.Context, event *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	event.CreatedAt = now
	event.UpdatedAt = now
	r.events[event.ID] = cloneEvent(event)
	return nil
}

func (r *EventRepository) Get(_ context.Context, id string) (*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	event, ok := r.events[id]
	if !ok {
		return nil, domain.ErrEventNotFound
	}
	return cloneEvent(event), nil
}

func (r *EventRepository) List(_ context.Context, filter domain.EventFilter) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]*domain.Event, 0, len(r.events))
	for _, event := range r.events {
		if matchesFilter(event, filter) {
			events = append(events, cloneEvent(event))
		}
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].StartTime.Equal(events[j].StartTime) {
			return events[i].CreatedAt.Before(events[j].CreatedAt)
		}
		return events[i].StartTime.Before(events[j].StartTime)
	})

	if filter.Offset >= len(events) {
		return []*domain.Event{}, nil
	}
	end := len(events)
	if filter.Limit > 0 && filter.Offset+filter.Limit < end {
		end = filter.Offset + filter.Limit
	}
	return events[filter.Offset:end], nil
}

func (r *EventRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.events[id]; !ok {
		return domain.ErrEventNotFound
	}
	delete(r.events, id)
	return nil
}

func matchesFilter(event *domain.Event, filter domain.EventFilter) bool {
	if filter.Sport != "" && !strings.EqualFold(event.Sport, filter.Sport) {
		return false
	}
	if filter.Category != "" && !strings.EqualFold(event.Category, filter.Category) {
		return false
	}
	if filter.Competition != "" && !strings.EqualFold(event.Competition, filter.Competition) {
		return false
	}
	if filter.Status != "" && !strings.EqualFold(event.Status, filter.Status) {
		return false
	}
	if filter.Country != "" && !strings.EqualFold(event.Country, filter.Country) {
		return false
	}
	if filter.City != "" && !strings.EqualFold(event.City, filter.City) {
		return false
	}
	if filter.Tag != "" && !hasValue(event.Tags, filter.Tag) {
		return false
	}
	return true
}

func hasValue(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func cloneEvent(event *domain.Event) *domain.Event {
	if event == nil {
		return nil
	}
	clone := *event
	clone.Participants = append([]string(nil), event.Participants...)
	clone.Tags = append([]string(nil), event.Tags...)
	return &clone
}
