package domain

import "time"

type Event struct {
	ID          string
<<<<<<< Updated upstream
	Title       string
	Sport       string
	Venue       string
	ScheduledAt time.Time
	Capacity    int
=======
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
>>>>>>> Stashed changes
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

<<<<<<< Updated upstream
type Registration struct {
	ID        string
	EventID   string
	UserID    string
	Status    string
	CreatedAt time.Time
=======
type EventFilter struct {
	Sport         string
	Competition   string
	StartTimeFrom time.Time
	StartTimeTo   time.Time
	Country       string
	Limit         int
	Offset        int
>>>>>>> Stashed changes
}
