package grpc

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func ClientStubs(stubsFunc func(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error) func(ctx context.Context,
	method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := stubsFunc(ctx, method, req, reply, cc, invoker, opts...)
		return err
	}
}

func slowRequest(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	time.Sleep(20 * time.Millisecond)
	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}

// GlobalStub is test double and it is used when we cannot or don’t want to involve real server.
// Instead of the real server, we introduced a stub and defined what data should be returned.
// Example:
// func(ctx context.Context,
// 		method string, req, reply interface{}, cc *grpc.ClientConn,
// 		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
// 		if method == "/helloworld.Greeter/SayHello" {
// 			reply.(*pb.HelloReply).Message = "hello"
// 		}
// 		return nil
// }.
var GlobalStub func(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error
