package infra

import (
	"code.jshyjdtech.com/godev/hykit/log"
	"code.jshyjdtech.com/godev/hykit/mysql"
	"gorm.io/gorm"
	"sync"

	"code.jshyjdtech.com/godev/hykit/redis"

	"code.jshyjdtech.com/godev/hykit/container"
	"code.jshyjdtech.com/godev/hykit/grpc"
	"github.com/google/wire"
)

// Do not change the function name and var name
//  infraOnce
//  onceInfra
//  infraSet
//  NewInfra

var (
	infraOnce sync.Once
	onceInfra *Infra
)

type Infra struct {
	*container.Esim

	DB *mysql.Client

	RedisClient *redis.Client
}

//nolint:deadcode,unused,varcheck
var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideDb,
	provideRedis,
)

func NewInfra() *Infra {
	infraOnce.Do(func() {
		esim := container.NewEsim()
		onceInfra = initInfra(esim, provideGrpcClient(esim))
	})

	return onceInfra
}

func NewStubsInfra(grpcClient *grpc.Client) *Infra {
	infraOnce.Do(func() {
		esim := container.NewEsim()
		onceInfra = initInfra(esim, grpcClient)
	})

	return onceInfra
}

// Close close the infra when app stop
func (infraer *Infra) Close() {
	// infraer.DB.Close()
}

func provideDb(esim *container.Esim) *mysql.Client {
	clientOptions := mysql.ClientOptions{}
	logger := log.NewGormLogger(
		log.WithGLogEsimZap(esim.Z),
	)

	mysqlClient := mysql.NewClient(
		clientOptions.WithConf(esim.Conf),
		clientOptions.WithLogger(esim.Logger),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: logger,
		}),
	)

	return mysqlClient
}

func provideGrpcClient(esim *container.Esim) *grpc.Client {
	clientOpt := grpc.ClientOption{}
	grpcClient := grpc.NewClient(
		clientOpt.WithLogger(esim.Logger),
		clientOpt.WithConf(esim.Conf),
	)
	return grpcClient
}

func provideRedis(esim *container.Esim) *redis.Client {
	clientOptions := redis.ClientOptions{}
	redisClent := redis.NewClient(
		clientOptions.WithConf(esim.Conf),
		clientOptions.WithLogger(esim.Logger),
		clientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := redis.MonitorProxyOptions{}
				return redis.NewMonitorProxy(
					monitorProxyOptions.WithConf(esim.Conf),
					monitorProxyOptions.WithLogger(esim.Logger),
					monitorProxyOptions.WithTracer(esim.Tracer),
				)
			},
		),
	)

	return redisClent
}
