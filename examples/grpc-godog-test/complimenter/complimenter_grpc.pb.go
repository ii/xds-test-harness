// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package complimenter

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

// ComplimenterClient is the client API for Complimenter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ComplimenterClient interface {
	GiveCompliment(ctx context.Context, in *ComplimentRequest, opts ...grpc.CallOption) (*ComplimentResponse, error)
}

type complimenterClient struct {
	cc grpc.ClientConnInterface
}

func NewComplimenterClient(cc grpc.ClientConnInterface) ComplimenterClient {
	return &complimenterClient{cc}
}

func (c *complimenterClient) GiveCompliment(ctx context.Context, in *ComplimentRequest, opts ...grpc.CallOption) (*ComplimentResponse, error) {
	out := new(ComplimentResponse)
	err := c.cc.Invoke(ctx, "/complimenter.Complimenter/GiveCompliment", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ComplimenterServer is the server API for Complimenter service.
// All implementations must embed UnimplementedComplimenterServer
// for forward compatibility
type ComplimenterServer interface {
	GiveCompliment(context.Context, *ComplimentRequest) (*ComplimentResponse, error)
	mustEmbedUnimplementedComplimenterServer()
}

// UnimplementedComplimenterServer must be embedded to have forward compatible implementations.
type UnimplementedComplimenterServer struct {
}

func (UnimplementedComplimenterServer) GiveCompliment(context.Context, *ComplimentRequest) (*ComplimentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GiveCompliment not implemented")
}
func (UnimplementedComplimenterServer) mustEmbedUnimplementedComplimenterServer() {}

// UnsafeComplimenterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ComplimenterServer will
// result in compilation errors.
type UnsafeComplimenterServer interface {
	mustEmbedUnimplementedComplimenterServer()
}

func RegisterComplimenterServer(s grpc.ServiceRegistrar, srv ComplimenterServer) {
	s.RegisterService(&Complimenter_ServiceDesc, srv)
}

func _Complimenter_GiveCompliment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ComplimentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ComplimenterServer).GiveCompliment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/complimenter.Complimenter/GiveCompliment",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ComplimenterServer).GiveCompliment(ctx, req.(*ComplimentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Complimenter_ServiceDesc is the grpc.ServiceDesc for Complimenter service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Complimenter_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "complimenter.Complimenter",
	HandlerType: (*ComplimenterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GiveCompliment",
			Handler:    _Complimenter_GiveCompliment_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "complimenter/complimenter.proto",
}