package utils

import (
	"batch_create/internal"
	"code.jshyjdtech.com/godev/hykit/log"
)

var GlobalLogger log.Logger

func init() {
	app := internal.NewApp()
	GlobalLogger = app.Logger
}
