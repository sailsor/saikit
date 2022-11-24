//go:build wireinject
// +build wireinject

package infra

import (
	"code.jshyjdtech.com/godev/hykit/container"
	"code.jshyjdtech.com/godev/hykit/grpc"
	"github.com/google/wire"
)

func initInfra(esim *container.Esim, grpc *grpc.Client) *Infra {
	wire.Build(infraSet)
	return nil
}
