package mysql

import (
	"gorm.io/gorm"
)

type ConnPool interface {
	// ConnPool 连接池
	gorm.ConnPool

	// ConnPoolBeginner 事物
	gorm.ConnPoolBeginner
}

type SQLClose interface {
	Close() error
}
