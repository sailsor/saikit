//go:build wireinject
// +build wireinject

package controllers

import (
	notify "notify/internal"

	"github.com/google/wire"
)

func initControllers(app *notify.App) *Controllers {
	wire.Build(controllersSet)
	return nil
}
