package authv1

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const AuthService_ServiceDescName = "sports.auth.v1.AuthService"

type AuthServiceClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*AuthResponse, error)
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*AuthResponse, error)
	ValidateToken(ctx context.Context, in *ValidateTokenRequest, opts ...grpc.CallOption) (*ValidateTokenResponse, error)
	GetProfile(ctx context.Context, in *GetProfileRequest, opts ...grpc.CallOption) (*UserProfileResponse, error)
}

type authServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthServiceClient(cc grpc.ClientConnInterface) AuthServiceClient {
	return &authServiceClient{cc}
}

func (c *authServiceClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*AuthResponse, error) {
	out := new(AuthResponse)
	err := c.cc.Invoke(ctx, "/sports.auth.v1.AuthService/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*AuthResponse, error) {
	out := new(AuthResponse)
	err := c.cc.Invoke(ctx, "/sports.auth.v1.AuthService/Login", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) ValidateToken(ctx context.Context, in *ValidateTokenRequest, opts ...grpc.CallOption) (*ValidateTokenResponse, error) {
	out := new(ValidateTokenResponse)
	err := c.cc.Invoke(ctx, "/sports.auth.v1.AuthService/ValidateToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) GetProfile(ctx context.Context, in *GetProfileRequest, opts ...grpc.CallOption) (*UserProfileResponse, error) {
	out := new(UserProfileResponse)
	err := c.cc.Invoke(ctx, "/sports.auth.v1.AuthService/GetProfile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type AuthServiceServer interface {
	Register(context.Context, *RegisterRequest) (*AuthResponse, error)
	Login(context.Context, *LoginRequest) (*AuthResponse, error)
	ValidateToken(context.Context, *ValidateTokenRequest) (*ValidateTokenResponse, error)
	GetProfile(context.Context, *GetProfileRequest) (*UserProfileResponse, error)
}

type UnimplementedAuthServiceServer struct{}

func (UnimplementedAuthServiceServer) Register(context.Context, *RegisterRequest) (*AuthResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedAuthServiceServer) Login(context.Context, *LoginRequest) (*AuthResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedAuthServiceServer) ValidateToken(context.Context, *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ValidateToken not implemented")
}
func (UnimplementedAuthServiceServer) GetProfile(context.Context, *GetProfileRequest) (*UserProfileResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetProfile not implemented")
}

func RegisterAuthServiceServer(s grpc.ServiceRegistrar, srv AuthServiceServer) {
	s.RegisterService(&AuthService_ServiceDesc, srv)
}

func _AuthService_Register_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.auth.v1.AuthService/Register"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(AuthServiceServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_Login_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.auth.v1.AuthService/Login"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(AuthServiceServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_ValidateToken_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(ValidateTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).ValidateToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.auth.v1.AuthService/ValidateToken"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(AuthServiceServer).ValidateToken(ctx, req.(*ValidateTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_GetProfile_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(GetProfileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).GetProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/sports.auth.v1.AuthService/GetProfile"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(AuthServiceServer).GetProfile(ctx, req.(*GetProfileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var AuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: AuthService_ServiceDescName,
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Register", Handler: _AuthService_Register_Handler},
		{MethodName: "Login", Handler: _AuthService_Login_Handler},
		{MethodName: "ValidateToken", Handler: _AuthService_ValidateToken_Handler},
		{MethodName: "GetProfile", Handler: _AuthService_GetProfile_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "services/auth-service/proto/auth/v1/auth.proto",
}
