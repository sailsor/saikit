package dbm

import (
	"batch_create/internal"
	"code.jshyjdtech.com/godev/hykit/log"
	"code.jshyjdtech.com/godev/hykit/mysql"
	"gorm.io/gorm"
)

var db_client *mysql.Client

func init() {
	app := internal.NewApp()
	esim := app.Esim

	clientOptions := mysql.ClientOptions{}
	gormLogger := log.NewGormLogger(
		log.WithGLogEsimZap(esim.Z),
	)
	db_client = mysql.NewClient(
		clientOptions.WithConf(esim.Conf),
		clientOptions.WithLogger(esim.Logger),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: gormLogger, // 会覆盖gorm的debug
		}),
	)

	app.Logger.Infof("taskpool_max_count: %v", app.Conf.GetInt64("taskpool_max_count"))
}

func GetDBClient() *mysql.Client {
	return db_client
}

func CloseDB() {
	db_client.Close()
}
