package routers

import (
	"notify/internal/transports/http/controllers"
	"notify/internal/transports/http/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterGinServer(en *gin.Engine, ctl *controllers.Controllers) {

	en.Use(gin.Recovery(), middlewares.ReporterParams(ctl.App.Logger))

	en.GET("/ping", ctl.Ping.Ping)
	en.POST("/ping", ctl.Ping.Ping)

	en.POST("/notify", ctl.Call.Notify)                          // 异步通知, 响应200
	en.POST("/notify/success", ctl.Call.NotifyWithSUCCESS)       // 异步通知, 响应200, SUCCESS
	en.POST("/notify/ok", ctl.Call.NotifyWithOK)                 // 异步通知, 响应200, ok
	en.POST("/notify/502", ctl.Call.NotifyWith502)               // 异步通知, 响应502
	en.POST("/notify/UnionJSCallback", ctl.Call.UnionJSCallback) // 解析江苏银联异步通知, 响应200, SUCCESS

}
