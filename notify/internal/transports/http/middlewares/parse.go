package middlewares

import (
	"net/http/httputil"
	"notify/internal/infra"

	"github.com/gin-gonic/gin"
)

func ParseHeaderHandler(infra *infra.Infra) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 请求头token

		ctx.Next()
	}
}

func DebugHttpHandler(infra *infra.Infra) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取请求报文
		request, _ := httputil.DumpRequest(ctx.Request, true)
		infra.Logger.Debugc(ctx, "http 请求[%s]", string(request))
		ctx.Next()
	}
}
