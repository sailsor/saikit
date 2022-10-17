package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

var (
	logger log.Logger = log.NewLogger()

	tcpAddr = &net.TCPAddr{IP: net.ParseIP(address).To4(), Port: port}
)

const (
	address = "0.0.0.0"

	port = 50051

	isTest = "is test"

	callPanic = "call_panic"

	callNil = "call_nil"

	callPanicArr = "callPanciArr"

	esim = "esim"
)

func TestNewGrpcClient(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOpt := ClientOption{}

	ctx := context.Background()
	client := NewClient(
		clientOpt.WithLogger(logger),
		clientOpt.WithConf(memConfig))
	conn, err := client.DialContext(ctx, tcpAddr.String())
	if err != nil {
		return
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn.Conn())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: esim})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.NotEmpty(t, r.Message)
	}
}

func TestGrpcClientPool(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)
	memConfig.Set("grpc_client_check_slow", true)
	memConfig.Set("grpc_client_slow_time", 10)

	clientOpt := ClientOption{}

	ctx := context.Background()
	client := NewClient(
		clientOpt.WithLogger(logger),
		clientOpt.WithConf(memConfig))
	pool, err := client.LoadServerPool(ctx, "TEST", tcpAddr.String())
	if err != nil {
		logger.Errorc(ctx, "连接池建立失败[%s]", err)
		return
	}
	assert.Nil(t, err)

	conn, err := pool.Get()
	if err != nil {
		logger.Errorc(ctx, "获取rpc连接失败[%s]", err)
		return
	}
	assert.Nil(t, err)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c := pb.NewGreeterClient(conn.Value())
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: esim})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.NotEmpty(t, r.Message)
	}

	conn, err = pool.Get()
	if err != nil {
		logger.Errorc(ctx, "获取rpc连接失败[%s]", err)
		return
	}
	assert.Nil(t, err)

	c = pb.NewGreeterClient(conn.Value())
	r, err = c.SayHello(ctx, &pb.HelloRequest{Name: esim})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.NotEmpty(t, r.Message)
	}

	//select {}
}

func TestSlowClient(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)
	memConfig.Set("grpc_client_check_slow", true)
	memConfig.Set("grpc_client_slow_time", 10)

	clientOpt := ClientOption{}

	ctx := context.Background()
	client := NewClient(
		clientOpt.WithLogger(logger),
		clientOpt.WithConf(memConfig),
		clientOpt.WithDialOptions(
			grpc.WithBlock(),
			grpc.WithChainUnaryInterceptor(slowRequest),
		))
	conn, err := client.DialContext(ctx, tcpAddr.String())
	if err != nil {
		return
	}

	defer conn.Close()
	c := pb.NewGreeterClient(conn.Conn())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: esim})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.NotEmpty(t, r.Message)
	}
}

func TestServerPanic(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOpt := ClientOption{}

	ctx := context.Background()
	client := NewClient(clientOpt.WithLogger(logger),
		clientOpt.WithConf(memConfig))
	conn, err := client.DialContext(ctx, tcpAddr.String())
	if err != nil {
		return
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn.Conn())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: callPanic})
	assert.Error(t, err)
	assert.Nil(t, r)
}

func TestServerPanicArr(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOpt := ClientOption{}

	ctx := context.Background()
	client := NewClient(clientOpt.WithLogger(logger),
		clientOpt.WithConf(memConfig))
	conn, err := client.DialContext(ctx, tcpAddr.String())
	if err != nil {
		return
	}

	defer conn.Close()
	c := pb.NewGreeterClient(conn.Conn())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: callPanicArr})
	assert.Error(t, err)
	assert.Nil(t, r)
}

func TestSubsReply(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOpt := ClientOption{}

	ctx := context.Background()
	client := NewClient(clientOpt.WithLogger(logger),
		clientOpt.WithConf(memConfig),
		clientOpt.WithDialOptions(
			grpc.WithChainUnaryInterceptor(ClientStubs(func(ctx context.Context,
				method string, req, reply interface{}, cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				if method == "/helloworld.Greeter/SayHello" {
					reply.(*pb.HelloReply).Message = esim
				}
				return nil
			})),
		))
	conn, err := client.DialContext(ctx, tcpAddr.String())
	if err != nil {
		return
	}

	defer conn.Close()
	c := pb.NewGreeterClient(conn.Conn())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: esim})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.Equal(t, esim, r.Message)
	}
}

func TestGlobalSubs(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("grpc_client_debug", true)

	clientOpt := ClientOption{}
	GlobalStub = func(ctx context.Context,
		method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if method == "/helloworld.Greeter/SayHello" {
			reply.(*pb.HelloReply).Message = esim
		}
		return nil
	}

	ctx := context.Background()
	client := NewClient(
		clientOpt.WithLogger(logger),
		clientOpt.WithConf(memConfig),
	)
	conn, err := client.DialContext(ctx, tcpAddr.String())
	if err != nil {
		return
	}

	defer conn.Close()
	c := pb.NewGreeterClient(conn.Conn())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: esim})
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.Equal(t, esim, r.Message)
	}
}
