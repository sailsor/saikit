package dbm

type TblUnionGroupIp struct {
	Id        int64  `gorm:"column:id"`
	Ip        string `gorm:"column:ip"`
	Port      string `gorm:"column:port"`
	Policy    string `gorm:"column:policy"`
	Reserve   string `gorm:"column:reserve"`
	GroupName string `gorm:"column:group_name"`
}

func (t *TblUnionGroupIp) TableName() string {
	return "tbl_union_group_ip"
}
