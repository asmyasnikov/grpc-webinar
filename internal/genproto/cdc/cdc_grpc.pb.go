// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.17.3
// source: cdc.proto

package cdc

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

// CDCClient is the client API for CDC service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CDCClient interface {
	Listen(ctx context.Context, in *ListenRequest, opts ...grpc.CallOption) (CDC_ListenClient, error)
}

type cDCClient struct {
	cc grpc.ClientConnInterface
}

func NewCDCClient(cc grpc.ClientConnInterface) CDCClient {
	return &cDCClient{cc}
}

func (c *cDCClient) Listen(ctx context.Context, in *ListenRequest, opts ...grpc.CallOption) (CDC_ListenClient, error) {
	stream, err := c.cc.NewStream(ctx, &CDC_ServiceDesc.Streams[0], "/cdc.CDC/Listen", opts...)
	if err != nil {
		return nil, err
	}
	x := &cDCListenClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type CDC_ListenClient interface {
	Recv() (*ListenResponse, error)
	grpc.ClientStream
}

type cDCListenClient struct {
	grpc.ClientStream
}

func (x *cDCListenClient) Recv() (*ListenResponse, error) {
	m := new(ListenResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CDCServer is the server API for CDC service.
// All implementations must embed UnimplementedCDCServer
// for forward compatibility
type CDCServer interface {
	Listen(*ListenRequest, CDC_ListenServer) error
	mustEmbedUnimplementedCDCServer()
}

// UnimplementedCDCServer must be embedded to have forward compatible implementations.
type UnimplementedCDCServer struct {
}

func (UnimplementedCDCServer) Listen(*ListenRequest, CDC_ListenServer) error {
	return status.Errorf(codes.Unimplemented, "method Listen not implemented")
}
func (UnimplementedCDCServer) mustEmbedUnimplementedCDCServer() {}

// UnsafeCDCServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CDCServer will
// result in compilation errors.
type UnsafeCDCServer interface {
	mustEmbedUnimplementedCDCServer()
}

func RegisterCDCServer(s grpc.ServiceRegistrar, srv CDCServer) {
	s.RegisterService(&CDC_ServiceDesc, srv)
}

func _CDC_Listen_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListenRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CDCServer).Listen(m, &cDCListenServer{stream})
}

type CDC_ListenServer interface {
	Send(*ListenResponse) error
	grpc.ServerStream
}

type cDCListenServer struct {
	grpc.ServerStream
}

func (x *cDCListenServer) Send(m *ListenResponse) error {
	return x.ServerStream.SendMsg(m)
}

// CDC_ServiceDesc is the grpc.ServiceDesc for CDC service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CDC_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "cdc.CDC",
	HandlerType: (*CDCServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Listen",
			Handler:       _CDC_Listen_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "cdc.proto",
}
