// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: remote_work_processor_service.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RemoteWorkProcessorServiceClient is the client API for RemoteWorkProcessorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RemoteWorkProcessorServiceClient interface {
	Session(ctx context.Context, opts ...grpc.CallOption) (RemoteWorkProcessorService_SessionClient, error)
}

type remoteWorkProcessorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRemoteWorkProcessorServiceClient(cc grpc.ClientConnInterface) RemoteWorkProcessorServiceClient {
	return &remoteWorkProcessorServiceClient{cc}
}

func (c *remoteWorkProcessorServiceClient) Session(ctx context.Context, opts ...grpc.CallOption) (RemoteWorkProcessorService_SessionClient, error) {
	stream, err := c.cc.NewStream(ctx, &RemoteWorkProcessorService_ServiceDesc.Streams[0], "/sap.autopilot.remote.work.processor.v1.RemoteWorkProcessorService/Session", opts...)
	if err != nil {
		return nil, err
	}
	x := &remoteWorkProcessorServiceSessionClient{stream}
	return x, nil
}

type RemoteWorkProcessorService_SessionClient interface {
	Send(*ClientMessage) error
	Recv() (*ServerMessage, error)
	grpc.ClientStream
}

type remoteWorkProcessorServiceSessionClient struct {
	grpc.ClientStream
}

func (x *remoteWorkProcessorServiceSessionClient) Send(m *ClientMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *remoteWorkProcessorServiceSessionClient) Recv() (*ServerMessage, error) {
	m := new(ServerMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RemoteWorkProcessorServiceServer is the server API for RemoteWorkProcessorService service.
// All implementations must embed UnimplementedRemoteWorkProcessorServiceServer
// for forward compatibility
type RemoteWorkProcessorServiceServer interface {
	Session(RemoteWorkProcessorService_SessionServer) error
	mustEmbedUnimplementedRemoteWorkProcessorServiceServer()
}

// UnimplementedRemoteWorkProcessorServiceServer must be embedded to have forward compatible implementations.
type UnimplementedRemoteWorkProcessorServiceServer struct {
}

func (UnimplementedRemoteWorkProcessorServiceServer) Session(RemoteWorkProcessorService_SessionServer) error {
	return status.Errorf(codes.Unimplemented, "method Session not implemented")
}
func (UnimplementedRemoteWorkProcessorServiceServer) mustEmbedUnimplementedRemoteWorkProcessorServiceServer() {
}

// UnsafeRemoteWorkProcessorServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RemoteWorkProcessorServiceServer will
// result in compilation errors.
type UnsafeRemoteWorkProcessorServiceServer interface {
	mustEmbedUnimplementedRemoteWorkProcessorServiceServer()
}

func RegisterRemoteWorkProcessorServiceServer(s grpc.ServiceRegistrar, srv RemoteWorkProcessorServiceServer) {
	s.RegisterService(&RemoteWorkProcessorService_ServiceDesc, srv)
}

func _RemoteWorkProcessorService_Session_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(RemoteWorkProcessorServiceServer).Session(&remoteWorkProcessorServiceSessionServer{stream})
}

type RemoteWorkProcessorService_SessionServer interface {
	Send(*ServerMessage) error
	Recv() (*ClientMessage, error)
	grpc.ServerStream
}

type remoteWorkProcessorServiceSessionServer struct {
	grpc.ServerStream
}

func (x *remoteWorkProcessorServiceSessionServer) Send(m *ServerMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *remoteWorkProcessorServiceSessionServer) Recv() (*ClientMessage, error) {
	m := new(ClientMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RemoteWorkProcessorService_ServiceDesc is the grpc.ServiceDesc for RemoteWorkProcessorService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RemoteWorkProcessorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "sap.autopilot.remote.work.processor.v1.RemoteWorkProcessorService",
	HandlerType: (*RemoteWorkProcessorServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Session",
			Handler:       _RemoteWorkProcessorService_Session_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "remote_work_processor_service.proto",
}
