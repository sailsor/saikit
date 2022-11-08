package utils

import (
	"code.jshyjdtech.com/godev/hykit/log"
	"gorm_tool/internal"
)

var GlobalLogger log.Logger

func InitGlobalLogger() {
	app := internal.NewApp()
	GlobalLogger = app.Logger
}
