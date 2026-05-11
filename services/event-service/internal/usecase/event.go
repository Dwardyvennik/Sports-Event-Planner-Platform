package usecase

import "context"

type EventRepository interface {
	Ping(context.Context) error
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
