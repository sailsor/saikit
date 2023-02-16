package main

import (
	"code.jshyjdtech.com/godev/hykit/mysql"
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm_tool/internal"
)

var (
	test1Config = mysql.DbConfig{
		Db: "bat",
		//Dsn:     "cd_mgm_data:AppUn!@69834@tcp(127.0.0.1:9100)/credit_bat_db?charset=utf8&parseTime=True&loc=Local", // 测试
		//Dsn:     "cd_bat_data:2m@BatDta@tcp(127.0.0.1:7242)/credit_bat_db?charset=utf8&parseTime=True&loc=Local", // 生产贷记
		//Dsn:     "db_bat_data:1m@BatDta@tcp(127.0.0.1:7242)/debit_bat_db?charset=utf8&parseTime=True&loc=Local", // 生产借记
		MaxIdle: 10,
		MaxOpen: 100}
)

type Diction struct {
	ColumnName string `gorm:"column:column_name"`
	DataType   string `gorm:"column:data_type"`
	TableName  string `gorm:"column:table_name"`
}

func main() {
	clientOptions := mysql.ClientOptions{}
	app := internal.NewApp()
	esim := app.Esim
	logger := esim.Logger

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

	ctx := context.Background()

	dicList := make([]Diction, 0)
	tableList := []string{
		"tbl_app_inf",
		"tbl_mcht_allacct_inf",
		"tbl_mcht_inf",
		"tbl_prod_order_log"}

	db.Table("information_schema.COLUMNS").
		Select("column_name,data_type,table_name").
		Where("data_type = ?", "varchar").
		Where("table_name in (?)", tableList).
		Find(&dicList)

	wordList := []string{"䶮", "㛃", "𠅤", "𫓩", "𬜬", "頔", "㼆", "韡", "𣓃", "𤩊"}
	for _, word := range wordList {
		logger.Infof("开始校验字[%s]", word)
		for _, table := range dicList {
			dict := ""
			//logger.Infoc(ctx, "开始校验表[%s]的字段[%s]", table.TableName, table.ColumnName)
			db.Table(table.TableName).Select(fmt.Sprintf("%s", table.ColumnName)).
				Where(fmt.Sprintf("%s", table.ColumnName)+" like ? ", "%"+word+"%").Limit(1).Scan(&dict)
			if dict != "" {
				logger.Errorc(ctx, "[%s][%s]存在生僻字[%s]", table.TableName, table.ColumnName, dict)
			}
		}
	}
	logger.Infof("应用执行完毕，退出....")
}
