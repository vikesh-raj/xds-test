// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.2
// source: echo.proto

package echo

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	EchoServer_SayHello_FullMethodName       = "/echo.EchoServer/SayHello"
	EchoServer_SayHelloStream_FullMethodName = "/echo.EchoServer/SayHelloStream"
)

// EchoServerClient is the client API for EchoServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EchoServerClient interface {
	SayHello(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoReply, error)
	SayHelloStream(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[EchoReply], error)
}

type echoServerClient struct {
	cc grpc.ClientConnInterface
}

func NewEchoServerClient(cc grpc.ClientConnInterface) EchoServerClient {
	return &echoServerClient{cc}
}

func (c *echoServerClient) SayHello(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (*EchoReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EchoReply)
	err := c.cc.Invoke(ctx, EchoServer_SayHello_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *echoServerClient) SayHelloStream(ctx context.Context, in *EchoRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[EchoReply], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &EchoServer_ServiceDesc.Streams[0], EchoServer_SayHelloStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[EchoRequest, EchoReply]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type EchoServer_SayHelloStreamClient = grpc.ServerStreamingClient[EchoReply]

// EchoServerServer is the server API for EchoServer service.
// All implementations must embed UnimplementedEchoServerServer
// for forward compatibility.
type EchoServerServer interface {
	SayHello(context.Context, *EchoRequest) (*EchoReply, error)
	SayHelloStream(*EchoRequest, grpc.ServerStreamingServer[EchoReply]) error
	mustEmbedUnimplementedEchoServerServer()
}

// UnimplementedEchoServerServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedEchoServerServer struct{}

func (UnimplementedEchoServerServer) SayHello(context.Context, *EchoRequest) (*EchoReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SayHello not implemented")
}
func (UnimplementedEchoServerServer) SayHelloStream(*EchoRequest, grpc.ServerStreamingServer[EchoReply]) error {
	return status.Errorf(codes.Unimplemented, "method SayHelloStream not implemented")
}
func (UnimplementedEchoServerServer) mustEmbedUnimplementedEchoServerServer() {}
func (UnimplementedEchoServerServer) testEmbeddedByValue()                    {}

// UnsafeEchoServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EchoServerServer will
// result in compilation errors.
type UnsafeEchoServerServer interface {
	mustEmbedUnimplementedEchoServerServer()
}

func RegisterEchoServerServer(s grpc.ServiceRegistrar, srv EchoServerServer) {
	// If the following call pancis, it indicates UnimplementedEchoServerServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&EchoServer_ServiceDesc, srv)
}

func _EchoServer_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EchoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EchoServerServer).SayHello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EchoServer_SayHello_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EchoServerServer).SayHello(ctx, req.(*EchoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EchoServer_SayHelloStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(EchoRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EchoServerServer).SayHelloStream(m, &grpc.GenericServerStream[EchoRequest, EchoReply]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type EchoServer_SayHelloStreamServer = grpc.ServerStreamingServer[EchoReply]

// EchoServer_ServiceDesc is the grpc.ServiceDesc for EchoServer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EchoServer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "echo.EchoServer",
	HandlerType: (*EchoServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _EchoServer_SayHello_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SayHelloStream",
			Handler:       _EchoServer_SayHelloStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "echo.proto",
}
