package grpc

import (
	"sync"
	"time"

	"github.com/pkg/errors"

	"code.jshyjdtech.com/godev/hykit/opentracing"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/prometheus/client_golang/prometheus"

	"code.jshyjdtech.com/godev/hykit/grpc/pool"
	"google.golang.org/grpc/connectivity"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	"github.com/davecgh/go-spew/spew"
	ggp "github.com/grpc-ecosystem/go-grpc-prometheus"
	opentracing2 "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var onceClient sync.Once
var gRpcClient *Client

type Client struct {
	logger log.Logger
	conf   config.Config

	clientMetrics *ggp.ClientMetrics
	tracer        opentracing2.Tracer

	/*仅作为DialContext服务使用*/
	connOpts []grpc.DialOption

	/*DialContext && Pool */
	opts []grpc.DialOption

	/*服务连接池映射*/
	mu          sync.Mutex
	servPoolMap sync.Map
}

type Option func(c *Client)

/*
* 对于每个进程生成单独唯一的Client 实例
 */
func NewClient(options ...Option) *Client {
	onceClient.Do(func() {
		client := &Client{}

		for _, option := range options {
			option(client)
		}

		if client.logger == nil {
			client.logger = log.NewLogger()
		}

		if client.conf == nil {
			client.conf = config.NewMemConfig()
		}

		/* gRpc 单独作为客户端时配置的参数信息，仅用于获取 ClientConn
		   DialOption
		*/
		keepAliveClient := keepalive.ClientParameters{}
		ClientKpTime := client.conf.GetInt("grpc_client_kp_time")
		if ClientKpTime == 0 {
			ClientKpTime = 60
		}
		keepAliveClient.Time = time.Duration(ClientKpTime) * time.Second

		ClientKpTimeOut := client.conf.GetInt("grpc_client_kp_time_out")
		if ClientKpTimeOut == 0 {
			ClientKpTimeOut = 5
		}
		keepAliveClient.Timeout = time.Duration(ClientKpTimeOut) * time.Second
		keepAliveClient.PermitWithoutStream = client.conf.GetBool("grpc_client_permit_without_stream")

		client.connOpts = []grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithKeepaliveParams(keepAliveClient),
		}

		//测试桩代码
		if GlobalStub != nil {
			client.connOpts = append(client.connOpts, grpc.WithChainUnaryInterceptor(ClientStubs(GlobalStub)))
		}

		// DialOption
		var gRpcOpts = make([]grpc.DialOption, 0)
		// opentracing 追踪
		if client.conf.GetBool("grpc_client_trace") {
			if client.tracer == nil {
				client.tracer = opentracing.NewTracer("grpc_client", client.logger)
			}
			tracerInterceptor := otgrpc.OpenTracingClientInterceptor(client.tracer)
			gRpcOpts = append(gRpcOpts, grpc.WithChainUnaryInterceptor(tracerInterceptor))
		}

		// prometheus统计
		if client.conf.GetBool("grpc_client_metrics") {
			if client.clientMetrics == nil {
				ggp.EnableClientHandlingTimeHistogram(ggp.WithHistogramBuckets(prometheus.DefBuckets))
				client.clientMetrics = ggp.DefaultClientMetrics
			}
			gRpcOpts = append(gRpcOpts, grpc.WithChainUnaryInterceptor(client.clientMetrics.UnaryClientInterceptor()))
		}

		if client.conf.GetBool("grpc_client_check_slow") {
			gRpcOpts = append(gRpcOpts, grpc.WithChainUnaryInterceptor(client.checkClientSlow()))
		}

		if client.conf.GetBool("grpc_client_debug") {
			gRpcOpts = append(gRpcOpts, grpc.WithChainUnaryInterceptor(client.clientDebug()))
		}

		// 测试桩代码
		if GlobalStub != nil {
			gRpcOpts = append(gRpcOpts, grpc.WithChainUnaryInterceptor(ClientStubs(GlobalStub)))
		}

		// 传入的grpc DialOption
		client.opts = append(gRpcOpts, client.opts...)

		gRpcClient = client
	})

	return gRpcClient
}

type ClientOption struct{}

func (ClientOption) WithConf(conf config.Config) Option {
	return func(g *Client) {
		g.conf = conf
	}
}

func (ClientOption) WithLogger(logger log.Logger) Option {
	return func(g *Client) {
		g.logger = logger
	}
}

func (ClientOption) WithTracer(tracer opentracing2.Tracer) Option {
	return func(g *Client) {
		g.tracer = tracer
	}
}

func (ClientOption) WithMetrics(metrics *ggp.ClientMetrics) Option {
	return func(g *Client) {
		g.clientMetrics = metrics
	}
}

func (ClientOption) WithDialOptions(options ...grpc.DialOption) Option {
	return func(g *Client) {
		g.opts = options
	}
}

func (gc *Client) checkClientSlow() func(ctx context.Context,
	method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ClientSlowTime := gc.conf.GetInt64("grpc_client_slow_time")

		beginTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		endTime := time.Now()

		if ClientSlowTime != 0 {
			if endTime.Sub(beginTime) > time.Duration(ClientSlowTime)*time.Millisecond {
				gc.logger.Warnc(ctx, "slow client grpc_handle %s", method)
			}
		}
		return err
	}
}

func (gc *Client) clientDebug() func(ctx context.Context,
	method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beginTime := time.Now()
		gc.logger.Debugc(ctx, "Grpc client start %s : %s, req : %s",
			cc.Target(), method, spew.Sdump(req))

		err := invoker(ctx, method, req, reply, cc, opts...)

		endTime := time.Now()
		gc.logger.Debugc(ctx, "Grpc client end [%v] %s : %s, reply : %s",
			endTime.Sub(beginTime).String(), cc.Target(), method, spew.Sdump(reply))

		return err
	}
}

/*返回单独的连接封装*/
type ClientConn struct {
	conn   *grpc.ClientConn
	cancel context.CancelFunc
	logger log.Logger
}

func (gc *Client) DialContext(ctx context.Context, serverAddr string) (*ClientConn, error) {
	var cancel context.CancelFunc
	var err error
	cc := &ClientConn{
		logger: gc.logger,
	}

	ClientConnTimeOut := gc.conf.GetInt("grpc_client_conn_time_out")
	if ClientConnTimeOut == 0 {
		ClientConnTimeOut = 30 //默认RPC超时时间30秒
	}

	ctx, cancel = context.WithTimeout(ctx, time.Duration(ClientConnTimeOut)*time.Second)
	cc.cancel = cancel

	/*请求服务端地址*/
	opts := gc.connOpts
	opts = append(opts, gc.opts...)
	cc.conn, err = grpc.DialContext(ctx, serverAddr, opts...)
	if err != nil {
		gc.logger.Errorc(ctx, "DialContext 到[%s] 失败[%s]", err)
		return nil, err
	}
	return cc, nil
}

func (cc *ClientConn) Conn() *grpc.ClientConn {
	return cc.conn
}

func (cc *ClientConn) State() connectivity.State {
	return cc.conn.GetState()
}

func (cc *ClientConn) Close() {
	_ = cc.conn.Close()
	cc.cancel()
}

type Pool struct {
	serverName string
	serverAddr string
	pool       pool.Pool
}

func (gc *Client) NewPool(ctx context.Context, servName, servAddr string) (*Pool, error) {
	var err error
	gcp := &Pool{
		serverName: servName,
		serverAddr: servAddr,
	}

	gc.logger.Infoc(ctx, "开始建立到[%s]地址[%s]的连接池;", servName, servAddr)

	//连接池初始化参数
	poolOpt := pool.DefaultOptions
	poolOpt.Dial = func(address string) (*grpc.ClientConn, error) {
		ctx, cancel := context.WithTimeout(context.Background(), pool.DialTimeout)
		defer cancel()
		var gRpcOpts = []grpc.DialOption{
			grpc.WithInsecure(),
			//grpc.WithConnectParams(grpc.ConnectParams{Backoff: backoff.Config{MaxDelay: pool.BackoffMaxDelay}}),
			grpc.WithInitialWindowSize(pool.InitialWindowSize),
			grpc.WithInitialConnWindowSize(pool.InitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(pool.MaxSendMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(pool.MaxRecvMsgSize)),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                pool.KeepAliveTime,
				Timeout:             pool.KeepAliveTimeout,
				PermitWithoutStream: true,
			}),
		}
		gRpcOpts = append(gRpcOpts, gc.opts...)
		return grpc.DialContext(ctx, address, gRpcOpts...)
	}
	//创建gRPC pool 连接池
	gcp.pool, err = pool.New(servAddr, poolOpt)
	if err != nil {
		gc.logger.Errorc(ctx, "创建[%s]地址[%s]连接池失败;[%s]", servName, servAddr, err)
		return nil, err
	}

	gc.logger.Infoc(ctx, "创建[%s]地址[%s]连接池成功", servName, servAddr)

	return gcp, nil
}

/**
pool 通过连接池 Get()方法返回可用的连接；
连接使用完成以后调用
	Close()方法归还到池里；
*/
func (gcp *Pool) Get() (pool.Conn, error) {
	return gcp.pool.Get()
}

/**
Close() 关闭gRPC连接池；
*/
func (gcp *Pool) Close() error {
	return gcp.pool.Close()
}

/**
Status() 获取连接池状态；
*/
func (gcp *Pool) Status() string {
	return gcp.pool.Status()
}

/**
Address() 返回连接地址；
*/
func (gcp *Pool) Address() string {
	return gcp.serverAddr
}

/**
ServerName() 返回连接服务名称；
*/
func (gcp *Pool) ServerName() string {
	return gcp.serverName
}

/*
	根据服务名称：服务目的地址初始化连接池，后续通过连接池进行相关操作；
	如果服务连接池未初始化,则通过相关参数初始化
	如果服务连接池已初始化，则直接返回连接池
*/
func (gc *Client) LoadServerPool(ctx context.Context, servName, servAddr string) (*Pool, error) {
	gc.logger.Infoc(ctx, "LoadServerPool 获取[%s]地址[%s]连接池;", servName, servAddr)

	if servName == "" || servAddr == "" {
		return nil, errors.Errorf("服务名称[%s]传入参数非法;", servName)
	}

	initPool := func(ctx context.Context) (*Pool, error) {
		p, err := gc.NewPool(ctx, servName, servAddr)
		if err != nil {
			gc.logger.Errorc(ctx, "LoadServerPool:NewPool[%s][%s]创建失败:[%s]", servName, servAddr, err)
			return nil, err
		}
		gc.logger.Infoc(ctx, "LoadServerPool:NewPool[%s][%s]创建成功；", servName, servAddr)
		return p, nil
	}

	//初始化或者获取已初始化Pool
	v, ok := gc.servPoolMap.Load(servName)
	if !ok {
		//不存在则创建新的链路
		gc.mu.Lock()
		//再次尝试获取
		v, ok = gc.servPoolMap.Load(servName)
		if ok { //匹配到
			if sPool, ok := v.(*Pool); ok {
				gc.mu.Unlock()
				return sPool, nil
			}
		}
		//创建连接池
		sPool, err := initPool(ctx)
		if err != nil {
			gc.logger.Errorc(ctx, "LoadServerPool:NewPool[%s][%s]创建失败:[%s]", servName, servAddr, err)
			gc.mu.Unlock()
			return nil, err
		}
		//设置服务连接池
		gc.servPoolMap.Store(servName, sPool)
		gc.mu.Unlock()
		return sPool, nil
	}
	//读取连接池
	if sPool, ok := v.(*Pool); ok {
		return sPool, nil
	} else {
		//重新创建连接池
		gc.mu.Lock()
		//创建连接池
		sPool, err := initPool(ctx)
		if err != nil {
			gc.logger.Errorc(ctx, "LoadServerPool:NewPool[%s][%s]创建失败:[%s]", servName, servAddr, err)
			gc.mu.Unlock()
			return nil, err
		}
		//设置服务连接池
		gc.servPoolMap.Store(servName, sPool)
		gc.mu.Unlock()
		return sPool, nil
	}
}

/*
	根据服务名称：更新服务连接池；
	服务连接池地址发生变化，则更新连接池信息，否则仍使用原连接池信息
*/
func (gc *Client) UpdateServerPool(ctx context.Context, servName, servAddr string) (*Pool, error) {
	gc.logger.Infoc(ctx, "UpdateServerPool 更新[%s]地址[%s]连接池;", servName, servAddr)
	if servName == "" || servAddr == "" {
		return nil, errors.Errorf("服务名称[%s]传入参数非法;", servName)
	}

	updatePool := func(ctx context.Context) (*Pool, error) {
		p, err := gc.NewPool(ctx, servName, servAddr)
		if err != nil {
			gc.logger.Errorc(ctx, "UpdateServerPool:NewPool[%s][%s]创建失败:[%s]", servName, servAddr, err)
			return nil, err
		}
		gc.logger.Infoc(ctx, "UpdateServerPool:NewPool[%s][%s]创建成功；", servName, servAddr)
		return p, nil
	}
	//获取服务信息
	v, ok := gc.servPoolMap.Load(servName)
	if !ok {
		return nil, errors.Errorf("UpdateServerPool[%s]不存在", servName)
	}
	//如原服务地址更新，则重新创建连接池
	sPool, ok := v.(*Pool)
	if !ok || sPool.serverAddr != servAddr {
		gc.mu.Lock()
		//重新测试
		v, ok = gc.servPoolMap.Load(servName)
		if ok && v.(*Pool).serverAddr == servAddr {
			gc.mu.Unlock()
			return v.(*Pool), nil
		}
		p, err := updatePool(ctx)
		if err != nil {
			gc.mu.Unlock()
			gc.logger.Errorc(ctx, "UpdateServerPool:initPool[%s][%s]初始化失败:[%s]", servName, servAddr, err)
			return nil, errors.Errorf("UpdateServerPool:[%s]连接池初始化失败[%s]", servName, err)
		}
		gc.servPoolMap.Store(servName, p)
		gc.mu.Unlock()
		gc.logger.Infoc(ctx, "UpdateServerPool:[%s]连接池更新成功；", servName)
		return p, nil
	}
	//如原连接池无更新信息或者状态正常，仍使用原连接池
	gc.logger.Infoc(ctx, "UpdateServerPool:[%s]连接池参数无变化，无需变更；", servName)
	return sPool, nil
}

/*
	根据服务名称：删除可用连接池；
*/
func (gc *Client) RemoveServerPool(ctx context.Context, servName string) error {
	gc.logger.Infoc(ctx, "RemoveServerPool 删除[%s]连接池;", servName)
	if servName == "" {
		return errors.Errorf("服务名称[%s]传入参数非法;", servName)
	}
	gc.servPoolMap.Delete(servName)
	return nil
}
