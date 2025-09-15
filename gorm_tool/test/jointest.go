package main

import (
	"encoding/json"
	"gorm_tool/dbm"
	"gorm_tool/utils"
)

type TradeInfo struct {
	ID int64 `json:"id,omitempty" gorm:"column:id;primary_key"`

	// 本地日期
	LocalDate string `json:"localDate" gorm:"column:local_date"`

	// 本地时间
	LocalTime string `json:"localTime" gorm:"column:local_time"`

	HostDate string `json:"hostDate" gorm:"column:host_date"`

	// 机构号
	AppID string `json:"appId" gorm:"column:app_id"`

	// 产品码
	ProdCd string `json:"prodCd" gorm:"column:prod_cd"`

	// 订单号
	OrderID string `json:"orderId" gorm:"column:order_id"`

	// 交易号
	TradeID string `json:"tradeId" gorm:"column:trade_id"`

	// 代扣渠道交易号
	TranSeq string `json:"tranSeq" gorm:"column:tran_seq"`

	// 交易状态
	TradeStatus string `json:"tradeStatus" gorm:"column:trade_status"`

	// 交易金额
	TranAmt string `json:"tranAmt" gorm:"column:tran_amt"`
}

func main() {
	TestJoin()
}

func TestJoin() {
	var err error
	logger := utils.GlobalLogger
	db := dbm.GetDBClient().GetDb("app_db").Table("tbl_prod_order_log as a")
	if db == nil {
		logger.Error("db is nil")
		return
	}

	tradeListInfo := make([]TradeInfo, 0)
	db = db.Debug().Where("a.id >= ?", 4081547)

	//err = db.Find(&tradeListInfo).Error
	//if err != nil {
	//	logger.Errorf("err: %s", err.Error())
	//	return
	//}

	err = db.Select("a.*, b.tran_seq").
		Joins("JOIN tbl_pay_log b on a.trade_id = b.tran_seq").
		Where("b.app_id = ?", "TEST_APP1").Debug().
		Find(&tradeListInfo).Error
	if err != nil {
		logger.Errorf("err: %s", err.Error())
		return
	}

	for _, trade := range tradeListInfo {
		msg, _ := json.Marshal(trade)
		logger.Infof("%s", msg)
	}
}
