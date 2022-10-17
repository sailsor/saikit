package grpc

import (
	"net"
	"runtime"
	"time"

	"code.jshyjdtech.com/godev/hykit/grpc/pool"

	"fmt"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	tracerid "code.jshyjdtech.com/godev/hykit/pkg/tracer-id"
	"github.com/davecgh/go-spew/spew"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	ggp "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Server *grpc.Server

	logger log.Logger

	conf config.Config

	unaryServerInterceptors []grpc.UnaryServerInterceptor

	opts []grpc.ServerOption

	target string

	serviceName string

	tracer opentracing2.Tracer
}

type ServerOption func(c *Server)

func NewServer(target string, options ...ServerOption) *Server {
	Server := &Server{}

	Server.target = target

	for _, option := range options {
		option(Server)
	}

	if Server.logger == nil {
		Server.logger = log.NewLogger()
	}

	if Server.conf == nil {
		Server.conf = config.NewNullConfig()
	}

	if Server.tracer == nil {
		Server.tracer = opentracing2.NoopTracer{}
	}

	unaryServerInterceptors := make([]grpc.UnaryServerInterceptor, 0)

	//trace
	if Server.conf.GetBool("grpc_server_trace") {
		unaryServerInterceptors = append(unaryServerInterceptors, otgrpc.OpenTracingServerInterceptor(Server.tracer))
		unaryServerInterceptors = append(unaryServerInterceptors, Server.tracerID())
	}

	if Server.conf.GetBool("grpc_server_metrics") {
		ggp.EnableHandlingTimeHistogram()
		serverMetrics := ggp.DefaultServerMetrics
		serverMetrics.EnableHandlingTimeHistogram(ggp.WithHistogramBuckets(prometheus.DefBuckets))
		unaryServerInterceptors = append(unaryServerInterceptors, serverMetrics.UnaryServerInterceptor())
	}

	if Server.conf.GetBool("grpc_server_check_slow") {
		unaryServerInterceptors = append(unaryServerInterceptors, Server.checkServerSlow())
	}

	if Server.conf.GetBool("grpc_server_debug") {
		unaryServerInterceptors = append(unaryServerInterceptors, Server.serverDebug())
	}

	// handle panic
	unaryServerInterceptors = append(unaryServerInterceptors, grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(Server.handelPanic())))

	if len(Server.unaryServerInterceptors) > 0 {
		unaryServerInterceptors = append(unaryServerInterceptors, Server.unaryServerInterceptors...)
	}

	var baseOpts = make([]grpc.ServerOption, 0)

	if len(unaryServerInterceptors) > 0 {
		ui := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryServerInterceptors...))
		baseOpts = append(baseOpts, ui)
	}

	if len(Server.opts) > 0 {
		baseOpts = append(baseOpts, Server.opts...)
	}

	//连接池优化参数
	baseOpts = append(baseOpts, grpc.InitialWindowSize(pool.InitialWindowSize),
		grpc.InitialConnWindowSize(pool.InitialConnWindowSize),
		grpc.MaxSendMsgSize(pool.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(pool.MaxRecvMsgSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:              pool.KeepAliveTime,
			Timeout:           pool.KeepAliveTimeout,
			MaxConnectionIdle: 5 * time.Minute,
		}))

	s := grpc.NewServer(baseOpts...)

	Server.Server = s

	return Server
}

type ServerOptions struct{}

func (ServerOptions) WithServerConf(conf config.Config) ServerOption {
	return func(g *Server) {
		g.conf = conf
	}
}

func (ServerOptions) WithServerLogger(logger log.Logger) ServerOption {
	return func(g *Server) {
		g.logger = logger
	}
}

func (ServerOptions) WithUnarySrvItcp(options ...grpc.UnaryServerInterceptor) ServerOption {
	return func(g *Server) {
		g.unaryServerInterceptors = options
	}
}

func (ServerOptions) WithServerOption(options ...grpc.ServerOption) ServerOption {
	return func(g *Server) {
		g.opts = options
	}
}

func (ServerOptions) WithTracer(tracer opentracing2.Tracer) ServerOption {
	return func(g *Server) {
		g.tracer = tracer
	}
}

func (gs *Server) checkServerSlow() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		beginTime := time.Now()
		resp, err = handler(ctx, req)
		endTime := time.Now()

		grpcClientSlowTime := gs.conf.GetInt64("grpc_server_slow_time")
		if grpcClientSlowTime != 0 {
			diffTime := endTime.Sub(beginTime)
			if endTime.Sub(beginTime) > time.Duration(grpcClientSlowTime)*time.Millisecond {
				gs.logger.Warnc(ctx, "Slow server %d %s", diffTime, info.FullMethod)
			}
		}

		return resp, err
	}
}

func (gs *Server) serverDebug() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		beginTime := time.Now()
		gs.logger.Debugc(ctx, "Grpc server start %s, req : %s", info.FullMethod, spew.Sdump(req))

		resp, err = handler(ctx, req)

		endTime := time.Now()
		gs.logger.Debugc(ctx, "Grpc server end [%v] %s, resp : %s",
			endTime.Sub(beginTime).String(),
			info.FullMethod, spew.Sdump(resp))

		return resp, err
	}
}

func (gs *Server) handelPanic() grpc_recovery.RecoveryHandlerFuncContext {
	return func(ctx context.Context, p interface{}) error {
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		gs.logger.Errorc(ctx, "panic recovered %s; [%s]", spew.Sdump(p), string(buf[:n]))
		return fmt.Errorf("server panic %v", p)
	}
}

// tracerId If not found opentracing's tracer_id then generate a new tracer_id.
func (gs *Server) tracerID() grpc.UnaryServerInterceptor {
	tracerID := tracerid.TracerID()
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if tracerid.ExtractTracerID(ctx) == "" {
			ctx = context.WithValue(ctx, tracerid.ActiveEsimKey, tracerID())
		}

		resp, err = handler(ctx, req)

		return resp, err
	}
}

//nolint:deadcode,unused
func nilResp() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		return nil, err
	}
}

func ServerStubs(stubsFunc func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error)) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		return stubsFunc(ctx, req, info, handler)
	}
}

func (gs *Server) Start() {
	lis, err := net.Listen("tcp", gs.target)
	if err != nil {
		gs.logger.Panicf("Failed to listen: %s", err.Error())
	}

	// Register reflection service on gRPC server.
	reflection.Register(gs.Server)

	gs.logger.Infof("Grpc server starting %s:%s",
		gs.serviceName, gs.target)
	go func() {
		if err := gs.Server.Serve(lis); err != nil {
			gs.logger.Panicf("Failed to start server: %s", err.Error())
		}
	}()
}

func (gs *Server) GracefulShutDown() {
	gs.Server.GracefulStop()
}
