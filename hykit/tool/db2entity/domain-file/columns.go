package domainfile

import (
	"strconv"
	"strings"

	"code.jshyjdtech.com/godev/hykit/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//nolint:deadcode,varcheck
const (
	golangByteArray  = "[]byte"
	gureguNullInt    = "null.Int"
	sqlNullInt       = "sql.NullInt64"
	golangInt        = "int"
	golangInt64      = "int64"
	gureguNullFloat  = "null.Float"
	sqlNullFloat     = "sql.NullFloat64"
	golangFloat      = "float"
	golangFloat32    = "float32"
	golangFloat64    = "float64"
	gureguNullString = "null.String"
	sqlNullString    = "sql.NullString"
	gureguNullTime   = "null.Time"
	golangTime       = "time.Time"
)

const (
	pri = "PRI"

	currentTimestamp = "CURRENT_TIMESTAMP"

	upCurrentTimestamp = "on update CURRENT_TIMESTAMP"

	yesNull = "YES"

	noNull = "NO"
)

type ColumnsRepo interface {
	SelectColumns(dbConf *DbConfig) (Columns, error)
}

type Column struct {
	ColumnName             string `gorm:"column:COLUMN_NAME"`
	ColumnKey              string `gorm:"column:COLUMN_KEY"`
	DataType               string `gorm:"column:DATA_TYPE"`
	IsNullAble             string `gorm:"column:IS_NULLABLE"`
	ColumnDefault          string `gorm:"column:COLUMN_DEFAULT"`
	CharacterMaximumLength string `gorm:"column:CHARACTER_MAXIMUM_LENGTH"`
	ColumnComment          string `gorm:"column:COLUMN_COMMENT"`
	Extra                  string `gorm:"column:EXTRA"`
}

type Columns []Column

func (cs Columns) Len() int {
	return len(cs)
}

func (cs Columns) IsEntity() bool {
	for i := range cs {
		if (&cs[i]).IsPri() {
			return true
		}
	}

	return false
}

type AutoTime struct {
	CurTimeStamp      []string
	OnUpdateTimeStamp []string
}

type DBColumnsInter struct {
	logger log.Logger
}

func NewDBColumnsInter(logger log.Logger) ColumnsRepo {
	dBColumnsInter := &DBColumnsInter{}
	dBColumnsInter.logger = logger
	return dBColumnsInter
}

// SelectColumns Select column details.
func (dc *DBColumnsInter) SelectColumns(dbConf *DbConfig) (Columns, error) {
	var err error
	var db *gorm.DB
	if dbConf.Password != "" {
		db, err = gorm.Open(mysql.Open(dbConf.User + ":" + dbConf.Password +
			"@tcp(" + dbConf.Host + ":" + strconv.Itoa(dbConf.Port) + ")/" +
			dbConf.Database + "?charset=utf8&parseTime=True&loc=Local"))
	} else {
		db, err = gorm.Open(mysql.Open(dbConf.User + "@tcp(" + dbConf.Host + ":" +
			strconv.Itoa(dbConf.Port) + ")/" + dbConf.Database + "?charset=utf8&parseTime=True&loc=Local"))
	}
	if err != nil {
		dc.logger.Panicf("Open mysql err: %s", err.Error())
	}

	db = db.Debug()
	dbc, err := db.DB()
	if err != nil {
		dc.logger.Panicf("Open mysql err: %s", err.Error())
	}
	defer dbc.Close()

	if !db.Migrator().HasTable(dbConf.Table) {
		dc.logger.Panicf("%s table not exists", dbConf.Table)
	}

	sql := "SELECT COLUMN_NAME, COLUMN_KEY, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, " +
		" CHARACTER_MAXIMUM_LENGTH, COLUMN_COMMENT, EXTRA " +
		"FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = ? AND table_name = ?"

	cs := make([]Column, 0)

	db.Raw(sql, dbConf.Database, dbConf.Table).Scan(&cs)

	if err != nil {
		dc.logger.Panicf(err.Error())
	}

	return cs, nil
}

//nolint:goconst
func (c *Column) GetGoType(nullable bool) string {
	switch c.DataType {
	case "tinyint", "int", "smallint", "mediumint":
		if nullable {
			return sqlNullInt
		}
		return golangInt
	case "bigint":
		if nullable {
			return sqlNullInt
		}
		return golangInt64
	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext":
		if nullable {
			return sqlNullString
		}
		return "string"
	case "date", "datetime", "time", "timestamp":
		return golangTime
	case "decimal", "double":
		if nullable {
			return sqlNullFloat
		}
		return golangFloat64
	case "float":
		if nullable {
			return sqlNullFloat
		}
		return golangFloat32
	case "binary", "blob", "longblob", "mediumblob", "varbinary":
		return golangByteArray
	}

	return ""
}

func (c *Column) CheckDelField() string {
	if strings.Contains(c.ColumnName, "del") &&
		strings.Contains(c.ColumnName, "is") {
		return c.ColumnName
	}

	return ""
}

func (c *Column) IsTime(goType string) bool {
	return goType == golangTime
}

func (c *Column) IsCurrentTimeStamp() bool {
	return c.ColumnDefault == currentTimestamp
}

func (c *Column) IsOnUpdate() bool {
	return c.Extra == upCurrentTimestamp
}

// filterComment 过滤和转义特殊字符.
func (c *Column) FilterComment() string {
	if c.ColumnComment != "" {
		c.ColumnComment = strings.Replace(c.ColumnComment, "\r", "\\r", -1)
		c.ColumnComment = strings.Replace(c.ColumnComment, "\n", "\\n", -1)
	}

	return c.ColumnComment
}

func (c *Column) IsPri() bool {
	return c.ColumnKey == pri
}

// GetDefCol 默认标签.
func (c *Column) GetDefCol() string {
	if c.ColumnDefault != currentTimestamp && c.ColumnDefault != "" {
		return ";default:" + c.ColumnDefault
		/*if strings.Contains(c.DataType, "int") {
			return ";default:" + c.ColumnDefault
		} else {
			return ";default:'" + c.ColumnDefault + "'"
		}*/
	}

	return ""
}
