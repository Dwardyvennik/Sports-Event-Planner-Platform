package grpc

import (
	"context"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/university/sports-event-planner-platform/services/event-service/internal/usecase"
	eventv1 "github.com/university/sports-event-planner-platform/services/event-service/proto/event/v1"
)

func Register(server *grpc.Server, events *usecase.EventUseCase, log *slog.Logger) {
	eventv1.RegisterEventServiceServer(server, &Server{
		events: events,
		log:    log,
	})
	log.Info("event grpc delivery registered")
}

type Server struct {
	eventv1.UnimplementedEventServiceServer
	events *usecase.EventUseCase
	log    *slog.Logger
}

func (s *Server) CreateEvent(ctx context.Context, req *eventv1.CreateEventRequest) (*eventv1.EventResponse, error) {
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Sport) == "" {
		return nil, status.Error(codes.InvalidArgument, "title and sport are required")
	}
	if err := s.events.Health(ctx); err != nil {
		return nil, status.Error(codes.Unavailable, "event dependencies unavailable")
	}
	return &eventv1.EventResponse{Event: eventFromCreate("event_stub", req)}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, req *eventv1.UpdateEventRequest) (*eventv1.EventResponse, error) {
	if strings.TrimSpace(req.EventId) == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}
	if err := s.events.Health(ctx); err != nil {
		return nil, status.Error(codes.Unavailable, "event dependencies unavailable")
	}
<<<<<<< Updated upstream
	return &eventv1.EventResponse{Event: &eventv1.Event{
		Id:          req.EventId,
		Title:       req.Title,
		Sport:       req.Sport,
		Venue:       req.Venue,
		ScheduledAt: req.ScheduledAt,
		Capacity:    req.Capacity,
	}}, nil
=======

	event, err := s.events.CreateEvent(ctx, usecase.CreateEventInput{
		Sport:       req.Sport,
		Category:    req.Category,
		Competition: req.Competition,
		Title:       req.Title,
		Description: req.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      req.Status,
		Country:     req.Country,
		City:        req.City,
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return &eventv1.EventResponse{Event: eventToProto(event)}, nil
>>>>>>> Stashed changes
}

func (s *Server) GetEvent(ctx context.Context, req *eventv1.GetEventRequest) (*eventv1.EventResponse, error) {
	if strings.TrimSpace(req.EventId) == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}
	if err := s.events.Health(ctx); err != nil {
		return nil, status.Error(codes.Unavailable, "event dependencies unavailable")
	}
	return &eventv1.EventResponse{Event: sampleEvent(req.EventId)}, nil
}

func (s *Server) ListEvents(ctx context.Context, req *eventv1.ListEventsRequest) (*eventv1.ListEventsResponse, error) {
<<<<<<< Updated upstream
	if err := s.events.Health(ctx); err != nil {
		return nil, status.Error(codes.Unavailable, "event dependencies unavailable")
=======
	startTimeFrom, err := parseOptionalTime(req.StartTimeFrom)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "start_time_from must be RFC3339")
	}
	startTimeTo, err := parseOptionalTime(req.StartTimeTo)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "start_time_to must be RFC3339")
	}

	events, err := s.events.ListEvents(ctx, usecase.ListEventsInput{
		Sport:         req.Sport,
		Competition:   req.Competition,
		StartTimeFrom: startTimeFrom,
		StartTimeTo:   startTimeTo,
		Country:       req.Country,
		Page:          int(req.Page),
		PageSize:      int(req.PageSize),
	})
	if err != nil {
		return nil, grpcError(err)
>>>>>>> Stashed changes
	}
	return &eventv1.ListEventsResponse{Events: []*eventv1.Event{
		sampleEvent("event_1"),
		{Id: "event_2", Title: "Campus Basketball Night", Sport: fallback(req.Sport, "basketball"), Venue: "Main Gym", ScheduledAt: "2026-06-05T18:00:00Z", Capacity: 80},
	}}, nil
}

func (s *Server) JoinEvent(_ context.Context, req *eventv1.JoinEventRequest) (*eventv1.EventActionResponse, error) {
	if strings.TrimSpace(req.EventId) == "" || strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id and user_id are required")
	}
	return &eventv1.EventActionResponse{Status: "joined"}, nil
}

func (s *Server) LeaveEvent(_ context.Context, req *eventv1.LeaveEventRequest) (*eventv1.EventActionResponse, error) {
	if strings.TrimSpace(req.EventId) == "" || strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id and user_id are required")
	}
	return &eventv1.EventActionResponse{Status: "left"}, nil
}

func (s *Server) GetUserEvents(ctx context.Context, req *eventv1.GetUserEventsRequest) (*eventv1.GetUserEventsResponse, error) {
	if strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if err := s.events.Health(ctx); err != nil {
		return nil, status.Error(codes.Unavailable, "event dependencies unavailable")
	}
	return &eventv1.GetUserEventsResponse{Events: []*eventv1.Event{sampleEvent("event_user_1")}}, nil
}

func eventFromCreate(id string, req *eventv1.CreateEventRequest) *eventv1.Event {
	return &eventv1.Event{
<<<<<<< Updated upstream
		Id:          id,
		Title:       req.Title,
		Sport:       req.Sport,
		Venue:       req.Venue,
		ScheduledAt: req.ScheduledAt,
		Capacity:    req.Capacity,
=======
		Id:          event.ID,
		Sport:       event.Sport,
		Category:    event.Category,
		Competition: event.Competition,
		Title:       event.Title,
		Description: event.Description,
		StartTime:   formatTime(event.StartTime),
		EndTime:     formatTime(event.EndTime),
		Status:      event.Status,
		Country:     event.Country,
		City:        event.City,
		CreatedAt:   formatTime(event.CreatedAt),
		UpdatedAt:   formatTime(event.UpdatedAt),
>>>>>>> Stashed changes
	}
}

func sampleEvent(id string) *eventv1.Event {
	return &eventv1.Event{
		Id:          id,
		Title:       "University Football Meetup",
		Sport:       "football",
		Venue:       "North Field",
		ScheduledAt: "2026-06-01T16:00:00Z",
		Capacity:    120,
	}
}

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}
