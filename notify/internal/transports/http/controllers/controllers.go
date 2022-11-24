package controllers

import (
	notify "notify/internal"
	"notify/internal/application"

	"github.com/google/wire"
)

type Controllers struct {
	App  *notify.App
	Ping *PingController
	Call *CallbackController
}

//nolint:deadcode,varcheck,unused
var controllersSet = wire.NewSet(
	wire.Struct(new(Controllers), "*"),
	providePingController,
	provideCallbackController,
)

func NewControllers(app *notify.App) *Controllers {
	controllers := initControllers(app)
	return controllers
}

func providePingController(app *notify.App) *PingController {
	pingController := &PingController{}
	return pingController
}

func provideCallbackController(app *notify.App) *CallbackController {
	Ctl := &CallbackController{}
	Ctl.Infra = app.Infra
	Ctl.UnApp = application.NewUnionJsCallbackSvc(app.Infra)
	return Ctl
}
