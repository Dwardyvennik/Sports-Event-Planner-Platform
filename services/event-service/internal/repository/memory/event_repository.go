package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dwardyvennik/sports-event-planner-platform/services/event-service/internal/domain"
)

type EventRepository struct {
	mu           sync.RWMutex
	events       map[string]*domain.Event
	participants map[string]map[string]struct{}
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		events:       make(map[string]*domain.Event),
		participants: make(map[string]map[string]struct{}),
	}
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
	return r.cloneWithCount(event), nil
}

func (r *EventRepository) List(_ context.Context, filter domain.EventFilter) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events := make([]*domain.Event, 0, len(r.events))
	for _, event := range r.events {
		if matchesFilter(event, filter) {
			events = append(events, r.cloneWithCount(event))
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

func (r *EventRepository) Update(_ context.Context, event *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.events[event.ID]
	if !ok {
		return domain.ErrEventNotFound
	}
	if existing.CreatorID != event.CreatorID {
		return domain.ErrEventForbidden
	}
	count := int32(len(r.participants[event.ID]))
	if event.MaxParticipants > 0 && count > event.MaxParticipants {
		return domain.ErrEventFull
	}
	event.CreatedAt = existing.CreatedAt
	event.UpdatedAt = time.Now().UTC()
	r.events[event.ID] = cloneEvent(event)
	return nil
}

func (r *EventRepository) Delete(_ context.Context, id string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	event, ok := r.events[id]
	if !ok {
		return domain.ErrEventNotFound
	}
	if event.CreatorID != userID {
		return domain.ErrEventForbidden
	}
	delete(r.events, id)
	delete(r.participants, id)
	return nil
}

func (r *EventRepository) Join(_ context.Context, id string, userID string) (*domain.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	event, ok := r.events[id]
	if !ok {
		return nil, domain.ErrEventNotFound
	}
	if r.participants[id] == nil {
		r.participants[id] = map[string]struct{}{}
	}
	if _, ok := r.participants[id][userID]; ok {
		return r.cloneWithCount(event), nil
	}
	if event.MaxParticipants > 0 && int32(len(r.participants[id])) >= event.MaxParticipants {
		return nil, domain.ErrEventFull
	}
	r.participants[id][userID] = struct{}{}
	return r.cloneWithCount(event), nil
}

func (r *EventRepository) Leave(_ context.Context, id string, userID string) (*domain.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	event, ok := r.events[id]
	if !ok {
		return nil, domain.ErrEventNotFound
	}
	delete(r.participants[id], userID)
	return r.cloneWithCount(event), nil
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
	return true
}

func (r *EventRepository) cloneWithCount(event *domain.Event) *domain.Event {
	clone := cloneEvent(event)
	clone.ParticipantsCount = int32(len(r.participants[event.ID]))
	return clone
}

func cloneEvent(event *domain.Event) *domain.Event {
	if event == nil {
		return nil
	}
	clone := *event
	return &clone
}
