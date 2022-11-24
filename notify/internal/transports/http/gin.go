package http

import (
	"context"
	"net/http"
	notify "notify/internal"
	"notify/internal/transports/http/controllers"
	"notify/internal/transports/http/routers"
	"strings"
	"time"

	"code.jshyjdtech.com/godev/hykit/log"
	middleware "code.jshyjdtech.com/godev/hykit/middle-ware"
	"github.com/gin-gonic/gin"
)

type GinServer struct {
	en *gin.Engine

	addr string

	logger log.Logger

	server *http.Server

	app *notify.App
}

func NewGinServer(app *notify.App) *GinServer {
	httpAddr := app.Conf.GetString("httpport")

	in := strings.Index(httpAddr, ":")
	if in < 0 {
		httpAddr = ":" + httpAddr
	}

	if app.Conf.GetString("runmode") != "pro" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	en := gin.Default()

	if app.Conf.GetBool("http_tracer") {
		en.Use(middleware.GinTracer(app.Tracer))
	}

	if app.Conf.GetBool("http_metrics") {
		en.Use(middleware.GinMonitor())
	}

	en.Use(middleware.GinTracerID(), gin.Recovery())

	server := &GinServer{
		en:     en,
		addr:   httpAddr,
		logger: app.Logger,
		app:    app,
	}

	return server
}

func (gs *GinServer) Start() {
	routers.RegisterGinServer(gs.en, controllers.NewControllers(gs.app))

	server := &http.Server{Addr: gs.addr, Handler: gs.en}
	gs.server = server
	gs.logger.Infof("gin start to listen %s", gs.addr)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				gs.logger.Fatalf("start http server err %s", err.Error())
			}
			return
		}
	}()
}

func (gs *GinServer) GracefulShutDown() {
	ctx, cannel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cannel()
	if err := gs.server.Shutdown(ctx); err != nil {
		gs.logger.Errorf("stop http server error %s", err.Error())
	}
}
