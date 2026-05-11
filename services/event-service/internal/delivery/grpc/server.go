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
	return &eventv1.EventResponse{Event: &eventv1.Event{
		Id:          req.EventId,
		Title:       req.Title,
		Sport:       req.Sport,
		Venue:       req.Venue,
		ScheduledAt: req.ScheduledAt,
		Capacity:    req.Capacity,
	}}, nil
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
	if err := s.events.Health(ctx); err != nil {
		return nil, status.Error(codes.Unavailable, "event dependencies unavailable")
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
		Id:          id,
		Title:       req.Title,
		Sport:       req.Sport,
		Venue:       req.Venue,
		ScheduledAt: req.ScheduledAt,
		Capacity:    req.Capacity,
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
