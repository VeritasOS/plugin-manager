// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.6.1
// source: pluginmanager/pm.proto

package pluginmanager

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

const (
	PM_Run_FullMethodName = "/pluginmanager.PM/Run"
)

// PMClient is the client API for PM service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PMClient interface {
	Run(ctx context.Context, in *RunRequest, opts ...grpc.CallOption) (*PluginTypeStatus, error)
}

type pMClient struct {
	cc grpc.ClientConnInterface
}

func NewPMClient(cc grpc.ClientConnInterface) PMClient {
	return &pMClient{cc}
}

func (c *pMClient) Run(ctx context.Context, in *RunRequest, opts ...grpc.CallOption) (*PluginTypeStatus, error) {
	out := new(PluginTypeStatus)
	err := c.cc.Invoke(ctx, PM_Run_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PMServer is the server API for PM service.
// All implementations must embed UnimplementedPMServer
// for forward compatibility
type PMServer interface {
	Run(context.Context, *RunRequest) (*PluginTypeStatus, error)
	mustEmbedUnimplementedPMServer()
}

// UnimplementedPMServer must be embedded to have forward compatible implementations.
type UnimplementedPMServer struct {
}

func (UnimplementedPMServer) Run(context.Context, *RunRequest) (*PluginTypeStatus, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Run not implemented")
}
func (UnimplementedPMServer) mustEmbedUnimplementedPMServer() {}

// UnsafePMServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PMServer will
// result in compilation errors.
type UnsafePMServer interface {
	mustEmbedUnimplementedPMServer()
}

func RegisterPMServer(s grpc.ServiceRegistrar, srv PMServer) {
	s.RegisterService(&PM_ServiceDesc, srv)
}

func _PM_Run_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RunRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PMServer).Run(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PM_Run_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PMServer).Run(ctx, req.(*RunRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PM_ServiceDesc is the grpc.ServiceDesc for PM service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PM_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pluginmanager.PM",
	HandlerType: (*PMServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Run",
			Handler:    _PM_Run_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pluginmanager/pm.proto",
}
