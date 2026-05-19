package eventv1

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const EventService_ServiceDescName = "sports.event.v1.EventService"

type EventServiceClient interface {
	CreateEvent(ctx context.Context, in *CreateEventRequest, opts ...grpc.CallOption) (*EventResponse, error)
	GetEvent(ctx context.Context, in *GetEventRequest, opts ...grpc.CallOption) (*EventResponse, error)
	ListEvents(ctx context.Context, in *ListEventsRequest, opts ...grpc.CallOption) (*ListEventsResponse, error)
	UpdateEvent(ctx context.Context, in *UpdateEventRequest, opts ...grpc.CallOption) (*EventResponse, error)
	DeleteEvent(ctx context.Context, in *DeleteEventRequest, opts ...grpc.CallOption) (*DeleteEventResponse, error)
	JoinEvent(ctx context.Context, in *EventMembershipRequest, opts ...grpc.CallOption) (*EventResponse, error)
	LeaveEvent(ctx context.Context, in *EventMembershipRequest, opts ...grpc.CallOption) (*EventResponse, error)
}

type eventServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewEventServiceClient(cc grpc.ClientConnInterface) EventServiceClient {
	return &eventServiceClient{cc}
}

func (c *eventServiceClient) CreateEvent(ctx context.Context, in *CreateEventRequest, opts ...grpc.CallOption) (*EventResponse, error) {
	out := new(EventResponse)
	err := c.cc.Invoke(ctx, "/sports.event.v1.EventService/CreateEvent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eventServiceClient) GetEvent(ctx context.Context, in *GetEventRequest, opts ...grpc.CallOption) (*EventResponse, error) {
	out := new(EventResponse)
	err := c.cc.Invoke(ctx, "/sports.event.v1.EventService/GetEvent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eventServiceClient) ListEvents(ctx context.Context, in *ListEventsRequest, opts ...grpc.CallOption) (*ListEventsResponse, error) {
	out := new(ListEventsResponse)
	err := c.cc.Invoke(ctx, "/sports.event.v1.EventService/ListEvents", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eventServiceClient) UpdateEvent(ctx context.Context, in *UpdateEventRequest, opts ...grpc.CallOption) (*EventResponse, error) {
	out := new(EventResponse)
	err := c.cc.Invoke(ctx, "/sports.event.v1.EventService/UpdateEvent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eventServiceClient) DeleteEvent(ctx context.Context, in *DeleteEventRequest, opts ...grpc.CallOption) (*DeleteEventResponse, error) {
	out := new(DeleteEventResponse)
	err := c.cc.Invoke(ctx, "/sports.event.v1.EventService/DeleteEvent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eventServiceClient) JoinEvent(ctx context.Context, in *EventMembershipRequest, opts ...grpc.CallOption) (*EventResponse, error) {
	out := new(EventResponse)
	err := c.cc.Invoke(ctx, "/sports.event.v1.EventService/JoinEvent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eventServiceClient) LeaveEvent(ctx context.Context, in *EventMembershipRequest, opts ...grpc.CallOption) (*EventResponse, error) {
	out := new(EventResponse)
	err := c.cc.Invoke(ctx, "/sports.event.v1.EventService/LeaveEvent", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type EventServiceServer interface {
	CreateEvent(context.Context, *CreateEventRequest) (*EventResponse, error)
	GetEvent(context.Context, *GetEventRequest) (*EventResponse, error)
	ListEvents(context.Context, *ListEventsRequest) (*ListEventsResponse, error)
	UpdateEvent(context.Context, *UpdateEventRequest) (*EventResponse, error)
	DeleteEvent(context.Context, *DeleteEventRequest) (*DeleteEventResponse, error)
	JoinEvent(context.Context, *EventMembershipRequest) (*EventResponse, error)
	LeaveEvent(context.Context, *EventMembershipRequest) (*EventResponse, error)
}

type UnimplementedEventServiceServer struct{}

func (UnimplementedEventServiceServer) CreateEvent(context.Context, *CreateEventRequest) (*EventResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateEvent not implemented")
}
func (UnimplementedEventServiceServer) GetEvent(context.Context, *GetEventRequest) (*EventResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetEvent not implemented")
}
func (UnimplementedEventServiceServer) ListEvents(context.Context, *ListEventsRequest) (*ListEventsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListEvents not implemented")
}
func (UnimplementedEventServiceServer) UpdateEvent(context.Context, *UpdateEventRequest) (*EventResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateEvent not implemented")
}
func (UnimplementedEventServiceServer) DeleteEvent(context.Context, *DeleteEventRequest) (*DeleteEventResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteEvent not implemented")
}
func (UnimplementedEventServiceServer) JoinEvent(context.Context, *EventMembershipRequest) (*EventResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method JoinEvent not implemented")
}
func (UnimplementedEventServiceServer) LeaveEvent(context.Context, *EventMembershipRequest) (*EventResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method LeaveEvent not implemented")
}

func RegisterEventServiceServer(s grpc.ServiceRegistrar, srv EventServiceServer) {
	s.RegisterService(&EventService_ServiceDesc, srv)
}

func _EventService_CreateEvent_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(CreateEventRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EventServiceServer).CreateEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.event.v1.EventService/CreateEvent"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(EventServiceServer).CreateEvent(ctx, req.(*CreateEventRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EventService_GetEvent_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(GetEventRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EventServiceServer).GetEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.event.v1.EventService/GetEvent"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(EventServiceServer).GetEvent(ctx, req.(*GetEventRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EventService_ListEvents_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(ListEventsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EventServiceServer).ListEvents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.event.v1.EventService/ListEvents"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(EventServiceServer).ListEvents(ctx, req.(*ListEventsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EventService_UpdateEvent_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(UpdateEventRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EventServiceServer).UpdateEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.event.v1.EventService/UpdateEvent"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(EventServiceServer).UpdateEvent(ctx, req.(*UpdateEventRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EventService_DeleteEvent_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(DeleteEventRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EventServiceServer).DeleteEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.event.v1.EventService/DeleteEvent"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(EventServiceServer).DeleteEvent(ctx, req.(*DeleteEventRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EventService_JoinEvent_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(EventMembershipRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EventServiceServer).JoinEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.event.v1.EventService/JoinEvent"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(EventServiceServer).JoinEvent(ctx, req.(*EventMembershipRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EventService_LeaveEvent_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(EventMembershipRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EventServiceServer).LeaveEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.event.v1.EventService/LeaveEvent"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(EventServiceServer).LeaveEvent(ctx, req.(*EventMembershipRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var EventService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: EventService_ServiceDescName,
	HandlerType: (*EventServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateEvent", Handler: _EventService_CreateEvent_Handler},
		{MethodName: "GetEvent", Handler: _EventService_GetEvent_Handler},
		{MethodName: "ListEvents", Handler: _EventService_ListEvents_Handler},
		{MethodName: "UpdateEvent", Handler: _EventService_UpdateEvent_Handler},
		{MethodName: "DeleteEvent", Handler: _EventService_DeleteEvent_Handler},
		{MethodName: "JoinEvent", Handler: _EventService_JoinEvent_Handler},
		{MethodName: "LeaveEvent", Handler: _EventService_LeaveEvent_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "services/event-service/proto/event/v1/event.proto",
}
