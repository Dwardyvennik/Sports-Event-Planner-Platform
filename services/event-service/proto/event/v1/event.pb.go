package eventv1

type Event struct {
	Id          string `json:"id,omitempty"`
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
}

type GetEventRequest struct {
	EventId string `json:"event_id,omitempty"`
}

type ListEventsRequest struct {
	Sport    string `json:"sport,omitempty"`
	Page     int32  `json:"page,omitempty"`
	PageSize int32  `json:"page_size,omitempty"`
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
