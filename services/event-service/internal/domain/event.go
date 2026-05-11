package domain

import "time"

type Event struct {
	ID          string
	Title       string
	Sport       string
	Venue       string
	ScheduledAt time.Time
	Capacity    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Registration struct {
	ID        string
	EventID   string
	UserID    string
	Status    string
	CreatedAt time.Time
}
