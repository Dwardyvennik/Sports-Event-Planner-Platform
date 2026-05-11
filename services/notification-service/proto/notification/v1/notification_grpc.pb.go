package notificationv1

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const NotificationService_ServiceDescName = "sports.notification.v1.NotificationService"

type NotificationServiceClient interface {
	SendNotification(ctx context.Context, in *SendNotificationRequest, opts ...grpc.CallOption) (*NotificationResponse, error)
	SendReminder(ctx context.Context, in *SendReminderRequest, opts ...grpc.CallOption) (*NotificationResponse, error)
	GetNotifications(ctx context.Context, in *GetNotificationsRequest, opts ...grpc.CallOption) (*GetNotificationsResponse, error)
}

type notificationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNotificationServiceClient(cc grpc.ClientConnInterface) NotificationServiceClient {
	return &notificationServiceClient{cc}
}

func (c *notificationServiceClient) SendNotification(ctx context.Context, in *SendNotificationRequest, opts ...grpc.CallOption) (*NotificationResponse, error) {
	out := new(NotificationResponse)
	err := c.cc.Invoke(ctx, "/sports.notification.v1.NotificationService/SendNotification", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) SendReminder(ctx context.Context, in *SendReminderRequest, opts ...grpc.CallOption) (*NotificationResponse, error) {
	out := new(NotificationResponse)
	err := c.cc.Invoke(ctx, "/sports.notification.v1.NotificationService/SendReminder", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *notificationServiceClient) GetNotifications(ctx context.Context, in *GetNotificationsRequest, opts ...grpc.CallOption) (*GetNotificationsResponse, error) {
	out := new(GetNotificationsResponse)
	err := c.cc.Invoke(ctx, "/sports.notification.v1.NotificationService/GetNotifications", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type NotificationServiceServer interface {
	SendNotification(context.Context, *SendNotificationRequest) (*NotificationResponse, error)
	SendReminder(context.Context, *SendReminderRequest) (*NotificationResponse, error)
	GetNotifications(context.Context, *GetNotificationsRequest) (*GetNotificationsResponse, error)
}

type UnimplementedNotificationServiceServer struct{}

func (UnimplementedNotificationServiceServer) SendNotification(context.Context, *SendNotificationRequest) (*NotificationResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method SendNotification not implemented")
}
func (UnimplementedNotificationServiceServer) SendReminder(context.Context, *SendReminderRequest) (*NotificationResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method SendReminder not implemented")
}
func (UnimplementedNotificationServiceServer) GetNotifications(context.Context, *GetNotificationsRequest) (*GetNotificationsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetNotifications not implemented")
}

func RegisterNotificationServiceServer(s grpc.ServiceRegistrar, srv NotificationServiceServer) {
	s.RegisterService(&NotificationService_ServiceDesc, srv)
}

func _NotificationService_SendNotification_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(SendNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).SendNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.notification.v1.NotificationService/SendNotification"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(NotificationServiceServer).SendNotification(ctx, req.(*SendNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_SendReminder_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(SendReminderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).SendReminder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.notification.v1.NotificationService/SendReminder"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(NotificationServiceServer).SendReminder(ctx, req.(*SendReminderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NotificationService_GetNotifications_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(GetNotificationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationServiceServer).GetNotifications(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.notification.v1.NotificationService/GetNotifications"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(NotificationServiceServer).GetNotifications(ctx, req.(*GetNotificationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var NotificationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: NotificationService_ServiceDescName,
	HandlerType: (*NotificationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "SendNotification", Handler: _NotificationService_SendNotification_Handler},
		{MethodName: "SendReminder", Handler: _NotificationService_SendReminder_Handler},
		{MethodName: "GetNotifications", Handler: _NotificationService_GetNotifications_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "services/notification-service/proto/notification/v1/notification.proto",
}
