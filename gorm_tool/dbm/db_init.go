package dbm

import (
	"code.jshyjdtech.com/godev/hykit/log"
	"code.jshyjdtech.com/godev/hykit/mysql"
	"gorm.io/gorm"
	"gorm_tool/internal"
)

var dbClient *mysql.Client

func init() {
	InitDB()
}

func InitDB() {
	app := internal.NewApp()
	esim := app.Esim

	clientOptions := mysql.ClientOptions{}
	gormLogger := log.NewGormLogger(
		log.WithGLogEsimZap(esim.Z),
	)
	dbClient = mysql.NewClient(
		clientOptions.WithConf(esim.Conf),
		clientOptions.WithLogger(esim.Logger),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: gormLogger, // 会覆盖gorm的debug
		}),
	)

}

func GetDBClient() *mysql.Client {
	return dbClient
}

func CloseDB() {
	dbClient.Close()
}
