package gorm_tool

import (
	"code.jshyjdtech.com/godev/hykit/mysql"
	"gorm.io/gorm"
	"gorm_tool/internal"
	"gorm_tool/utils"
	"testing"
)

var DBClient *mysql.Client

func init() {
	clientOptions := mysql.ClientOptions{}
	app := internal.NewApp()
	esim := app.Esim

	DBClient = mysql.NewClient(
		clientOptions.WithConf(esim.Conf),
		clientOptions.WithLogger(esim.Logger),
		clientOptions.WithGormConfig(&gorm.Config{
			//Logger: nil, //
		}),
	)
	utils.InitGlobalLogger()
}

type AuditLog struct {
	ID        int64
	UserName  string
	ClientIp  string
	OptMenu   string
	OptId     string
	OptName   string
	OptResult string
	LocalDate string
}

func TestGormNullWithEmpty(t *testing.T) {

	var err error
	c := DBClient
	logger := utils.GlobalLogger
	defer c.Close()

	db := c.GetDb("app_db")
	if db == nil {
		logger.Error("db is nil")
		return
	}

	list := make([]AuditLog, 0)

	err = db.Table("audit_log").Select("*").Limit(10).Find(&list).Error
	if err != nil {
		logger.Error("查询失败[%s]")
		return
	}

	for _, audit := range list {
		logger.Infof("%+v", audit)
	}
}
