package gorm_tool

import (
	"code.jshyjdtech.com/godev/hykit/log"
	"code.jshyjdtech.com/godev/hykit/mysql"
	"context"
	"gorm.io/gorm"
	"gorm_tool/internal"
	"testing"
)

func TestNewClient(t *testing.T) {

	app := internal.NewApp()
	esim := app.Esim

	clientOptions := mysql.ClientOptions{}
	gormLogger := log.NewGormLogger(
		log.WithGLogEsimZap(esim.Z),
	)
	gormLogger.Info(context.Background(), "hello")
	c := mysql.NewClient(
		clientOptions.WithConf(esim.Conf),
		clientOptions.WithLogger(esim.Logger),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: gormLogger, // 会覆盖gorm的debug
		}),
	)
	defer c.Close()

	db := c.GetDb("app_db")
	db.Table("user_inf").Select("*").Limit(1).Find(&UserInf{})
}

type UserInf struct {
	Id     int64
	Name   string
	Salary string
	Age    int64
}

var (
	test1Config = mysql.DbConfig{
		Db:      "bat",
		Dsn:     "goesim:goesim@12345678@tcp(rm-bp11vuqb6wz9476nbym.mysql.rds.aliyuncs.com:3306)/test_db?charset=utf8&parseTime=True&loc=Local",
		MaxIdle: 10,
		MaxOpen: 100}
)

func TestNewClientWithDbConfig(t *testing.T) {
	clientOptions := mysql.ClientOptions{}
	app := internal.NewApp()
	esim := app.Esim

	c := mysql.NewClient(
		clientOptions.WithDbConfig([]mysql.DbConfig{test1Config}),
		clientOptions.WithLogger(esim.Logger),
		clientOptions.WithGormConfig(&gorm.Config{
			//Logger: nil, //
		}),
	)
	defer c.Close()

	db := c.GetDb("bat")
	if db == nil {
		esim.Logger.Error("db is nil")
		return
	}
	db.Table("user_inf").Select("*").Limit(1).Find(&UserInf{})
}
