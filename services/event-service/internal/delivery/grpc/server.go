package grpc

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/university/sports-event-planner-platform/services/event-service/internal/domain"
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
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Sport) == "" || strings.TrimSpace(req.StartTime) == "" {
		return nil, status.Error(codes.InvalidArgument, "title, sport, and start_time are required")
	}

	startTime, err := parseTime(req.StartTime)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "start_time must be RFC3339")
	}
	endTime, err := parseOptionalTime(req.EndTime)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "end_time must be RFC3339")
	}

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
}

func (s *Server) GetEvent(ctx context.Context, req *eventv1.GetEventRequest) (*eventv1.EventResponse, error) {
	event, err := s.events.GetEvent(ctx, req.EventId)
	if err != nil {
		return nil, grpcError(err)
	}
	return &eventv1.EventResponse{Event: eventToProto(event)}, nil
}

func (s *Server) ListEvents(ctx context.Context, req *eventv1.ListEventsRequest) (*eventv1.ListEventsResponse, error) {
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
	}

	response := &eventv1.ListEventsResponse{Events: make([]*eventv1.Event, 0, len(events))}
	for _, event := range events {
		response.Events = append(response.Events, eventToProto(event))
	}
	return response, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *eventv1.DeleteEventRequest) (*eventv1.DeleteEventResponse, error) {
	if err := s.events.DeleteEvent(ctx, req.EventId); err != nil {
		return nil, grpcError(err)
	}
	return &eventv1.DeleteEventResponse{EventId: strings.TrimSpace(req.EventId), Deleted: true}, nil
}

func eventToProto(event *domain.Event) *eventv1.Event {
	if event == nil {
		return nil
	}
	return &eventv1.Event{
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
	}
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(value))
}

func parseOptionalTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	return parseTime(value)
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func grpcError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidEvent):
		return status.Error(codes.InvalidArgument, "invalid event")
	case errors.Is(err, domain.ErrEventNotFound):
		return status.Error(codes.NotFound, "event not found")
	default:
		return status.Error(codes.Internal, "event service error")
	}
}
