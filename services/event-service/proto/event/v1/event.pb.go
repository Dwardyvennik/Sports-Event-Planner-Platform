package eventv1

type Event struct {
	Id          string `json:"id,omitempty"`
<<<<<<< Updated upstream
	Title       string `json:"title,omitempty"`
	Sport       string `json:"sport,omitempty"`
	Venue       string `json:"venue,omitempty"`
	ScheduledAt string `json:"scheduled_at,omitempty"`
	Capacity    int32  `json:"capacity,omitempty"`
}

type CreateEventRequest struct {
	Title       string `json:"title,omitempty"`
	Sport       string `json:"sport,omitempty"`
	Venue       string `json:"venue,omitempty"`
	ScheduledAt string `json:"scheduled_at,omitempty"`
	Capacity    int32  `json:"capacity,omitempty"`
}

type UpdateEventRequest struct {
	EventId     string `json:"event_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Sport       string `json:"sport,omitempty"`
	Venue       string `json:"venue,omitempty"`
	ScheduledAt string `json:"scheduled_at,omitempty"`
	Capacity    int32  `json:"capacity,omitempty"`
=======
	Sport       string `json:"sport,omitempty"`
	Category    string `json:"category,omitempty"`
	Competition string `json:"competition,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	StartTime   string `json:"start_time,omitempty"`
	EndTime     string `json:"end_time,omitempty"`
	Status      string `json:"status,omitempty"`
	Country     string `json:"country,omitempty"`
	City        string `json:"city,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

type CreateEventRequest struct {
	Sport       string `json:"sport,omitempty"`
	Category    string `json:"category,omitempty"`
	Competition string `json:"competition,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	StartTime   string `json:"start_time,omitempty"`
	EndTime     string `json:"end_time,omitempty"`
	Status      string `json:"status,omitempty"`
	Country     string `json:"country,omitempty"`
	City        string `json:"city,omitempty"`
>>>>>>> Stashed changes
}

type GetEventRequest struct {
	EventId string `json:"event_id,omitempty"`
}

type ListEventsRequest struct {
<<<<<<< Updated upstream
	Sport    string `json:"sport,omitempty"`
	Page     int32  `json:"page,omitempty"`
	PageSize int32  `json:"page_size,omitempty"`
=======
	Sport         string `json:"sport,omitempty"`
	Competition   string `json:"competition,omitempty"`
	StartTimeFrom string `json:"start_time_from,omitempty"`
	StartTimeTo   string `json:"start_time_to,omitempty"`
	Country       string `json:"country,omitempty"`
	Page          int32  `json:"page,omitempty"`
	PageSize      int32  `json:"page_size,omitempty"`
>>>>>>> Stashed changes
}

type EventResponse struct {
	Event *Event `json:"event,omitempty"`
}

type ListEventsResponse struct {
	Events []*Event `json:"events,omitempty"`
}

type JoinEventRequest struct {
	EventId string `json:"event_id,omitempty"`
	UserId  string `json:"user_id,omitempty"`
}

type LeaveEventRequest struct {
	EventId string `json:"event_id,omitempty"`
	UserId  string `json:"user_id,omitempty"`
}

type EventActionResponse struct {
	Status string `json:"status,omitempty"`
}

type GetUserEventsRequest struct {
	UserId string `json:"user_id,omitempty"`
}

type GetUserEventsResponse struct {
	Events []*Event `json:"events,omitempty"`
}
