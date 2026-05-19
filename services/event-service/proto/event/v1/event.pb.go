package eventv1

type Event struct {
	Id          string `json:"id,omitempty"`
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
}

type GetEventRequest struct {
	EventId string `json:"event_id,omitempty"`
}

type UpdateEventRequest struct {
	Id              string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Title           string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Sport           string `protobuf:"bytes,3,opt,name=sport,proto3" json:"sport,omitempty"`
	Location        string `protobuf:"bytes,4,opt,name=location,proto3" json:"location,omitempty"`
	StartTime       string `protobuf:"bytes,5,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	EndTime         string `protobuf:"bytes,6,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`
	MaxParticipants int32  `protobuf:"varint,7,opt,name=max_participants,json=maxParticipants,proto3" json:"max_participants,omitempty"`
}

type ListEventsRequest struct {
	Sport         string `json:"sport,omitempty"`
	Competition   string `json:"competition,omitempty"`
	StartTimeFrom string `json:"start_time_from,omitempty"`
	StartTimeTo   string `json:"start_time_to,omitempty"`
	Country       string `json:"country,omitempty"`
	Page          int32  `json:"page,omitempty"`
	PageSize      int32  `json:"page_size,omitempty"`
}

type EventResponse struct {
	Event *Event `json:"event,omitempty"`
}

type ListEventsResponse struct {
	Events []*Event `json:"events,omitempty"`
}

type DeleteEventRequest struct {
	EventId string `json:"event_id,omitempty"`
}

type DeleteEventResponse struct {
	EventId string `json:"event_id,omitempty"`
	Deleted bool   `json:"deleted,omitempty"`
}
