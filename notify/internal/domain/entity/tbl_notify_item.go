package entity

import "time"

type TblNotifyItem struct {
	Id        int64     `gorm:"column:id"`
	AppId     string    `gorm:"column:app_id"`
	OrderId   string    `gorm:"column:order_id"`
	TradeId   string    `gorm:"column:trade_id"`
	Status    string    `gorm:"column:status"`
	TranTime  string    `gorm:"column:tran_time"`
	CreatedAt time.Time `gorm:"column:created_at;default:'CURRENT_TIMESTAMP(6)'"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:'CURRENT_TIMESTAMP(6)'"`
}

func (TblNotifyItem) TableName() string {
	return "tbl_notify_item"
}
