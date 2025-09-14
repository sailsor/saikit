package internal

import (
	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/container"
	"code.jshyjdtech.com/godev/hykit/log"
	eot "code.jshyjdtech.com/godev/hykit/opentracing"
	"code.jshyjdtech.com/godev/hykit/prometheus"
	"os"
)

const defaultAppname = "esim"
const defaultPrometheusHTTPArrd = "9002"

type App struct {
	*container.Esim

	confPath []string
}

type Option func(c *App)

type AppOptions struct{}

func NewApp(options ...Option) *App {
	app := &App{}

	for _, option := range options {
		option(app)
	}

	if app.confPath == nil {
		app.confPath = []string{"conf/"}
	}

	confFile := "conf"

	confOps := config.ViperConfOptions{}
	conf := config.NewViperConfig(
		confOps.WithConfigType("yaml"),
		confOps.WithConfPath(app.confPath),
		confOps.WithConfFile([]string{confFile}))

	env := os.Getenv("ENV")
	if env == "" {
		conf.Set("runmode", "dev")
	}

	ez := log.NewEsimZap(
		log.WithEsimZapConf(conf),
		log.WithEsimZapDebug(conf.GetBool("debug")),
		log.WithEsimZapJSON(conf.GetString("runmode") == "pro"),
	)

	logger := log.NewLogger(
		log.WithEsimZap(ez),
	)

	appname := defaultAppname
	if conf.GetString("appname") != "" {
		appname = conf.GetString("appname")
	}
	tracer := eot.NewTracer(appname, logger)

	promer := prometheus.NewNullProme()

	app.Esim = container.NewEsim(
		container.WithEsimZap(ez),
		container.WithLogger(logger),
		container.WithConf(conf),
		container.WithTracer(tracer),
		container.WithPromer(promer),
	)

	return app
}

func (AppOptions) WithConfPath(confPath []string) Option {
	return func(app *App) {
		app.confPath = confPath
	}
}
