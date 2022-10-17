package example1

import (
	"context"

	aa "code.jshyjdtech.com/godev/hykit/redis"
	"code.jshyjdtech.com/godev/hykit/tool/ifacer/example/repo"
	"github.com/gomodule/redigo/redis"
)

type TestStruct struct {
}

type Close interface {
	Close(string, int) error

	Err() error
}

type Test interface {
	Iface1(func(string) string) (result bool, pool redis.Pool)

	Iface2(ctx context.Context, found *bool) (result bool, err error)

	Iface3() (f func(repo.Repo) string)

	Iface4(map[string]*aa.Client) map[string]string

	Iface5(redisClient *aa.Client) *aa.Client

	Iface6(redisClient aa.Client) aa.Client

	Iface7(chan<- bool, chan<- aa.Client) <-chan bool

	Iface8(rp repo.Repo) repo.Repo

	Close

	Iface9(TestStruct, []TestStruct, [3]TestStruct)

	Iface10(Close)

	Iface11(...interface{})
}
