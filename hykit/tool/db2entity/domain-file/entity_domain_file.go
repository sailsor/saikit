package domainfile

import (
	"fmt"
	"path/filepath"
	"strings"

	"code.jshyjdtech.com/godev/hykit/log"
	"code.jshyjdtech.com/godev/hykit/pkg"
	filedir "code.jshyjdtech.com/godev/hykit/pkg/file-dir"
	"code.jshyjdtech.com/godev/hykit/pkg/templates"
	"github.com/serenize/snaker"
	"github.com/spf13/viper"
)

type entityDomainFile struct {
	withBoubctx string

	withEntityTarget string

	withDisbleEntity bool

	name string

	template string

	data entityTpl

	logger log.Logger

	tpl templates.Tpl

	tableName string
}

type EntityDomainFileOption func(*entityDomainFile)

func NewEntityDomainFile(options ...EntityDomainFileOption) DomainFile {
	e := &entityDomainFile{}

	for _, option := range options {
		option(e)
	}

	e.name = "entity"

	e.template = entityTemplate

	return e
}

func WithEntityDomainFileLogger(logger log.Logger) EntityDomainFileOption {
	return func(e *entityDomainFile) {
		e.logger = logger
	}
}

func WithEntityDomainFileTpl(tpl templates.Tpl) EntityDomainFileOption {
	return func(e *entityDomainFile) {
		e.tpl = tpl
	}
}

// Disabled implements DomainFile.
// 不推荐关闭实体文件.
func (edf *entityDomainFile) Disabled() bool {
	return edf.withDisbleEntity
}

// bindInput implements DomainFile.
func (edf *entityDomainFile) BindInput(v *viper.Viper) error {
	boubctx := v.GetString("boubctx")
	if boubctx != "" {
		edf.withBoubctx = boubctx + string(filepath.Separator)
	}

	edf.withDisbleEntity = v.GetBool("disable_entity")
	if !edf.withDisbleEntity {
		edf.withEntityTarget = v.GetString("entity_target")

		if edf.withEntityTarget == "" {
			if edf.withBoubctx != "" {
				edf.withEntityTarget = "internal" + string(filepath.Separator) + "domain" +
					string(filepath.Separator) + edf.withBoubctx + "entity"
			} else {
				edf.withEntityTarget = "internal" + string(filepath.Separator) + "domain" +
					string(filepath.Separator) + "entity"
			}
		} else {
			edf.withEntityTarget = strings.TrimLeft(edf.withEntityTarget, "/")
			edf.withEntityTarget = edf.withBoubctx + edf.withEntityTarget
		}

		entityTargetExists, err := filedir.IsExistsDir(edf.withEntityTarget)
		if err != nil {
			return err
		}

		if !entityTargetExists {
			err = filedir.CreateDir(edf.withEntityTarget)
			if err != nil {
				return err
			}
		}

		edf.withEntityTarget += string(filepath.Separator)
	}

	edf.logger.Debugf("withEntityTarget %s", edf.withEntityTarget)

	return nil
}

// parseCloumns implements DomainFile.
func (edf *entityDomainFile) ParseCloumns(cs Columns, info *ShareInfo) {
	tpl := entityTpl{}

	if cs.Len() == 0 {
		return
	}

	edf.tableName = info.DbConf.Table
	tpl.StructName = info.CamelStruct

	structInfo := templates.NewStructInfo()

	var colDefault string
	var valueType string
	var doc string
	var nullable bool
	var fieldName string
	var delField string

	for i := range cs {
		column := (&cs[i])

		field := pkg.Field{}

		fieldName = snaker.SnakeToCamel(column.ColumnName)
		field.Name = fieldName

		if column.IsNullAble == "YES" {
			nullable = true
		}

		valueType = column.GetGoType(nullable)
		if column.IsTime(valueType) {
			tpl.Imports = append(tpl.Imports, pkg.Import{Path: "time"})
		} else if strings.Contains(valueType, "sql.") {
			tpl.Imports = append(tpl.Imports, pkg.Import{Path: "database/sql"})
		}
		field.Type = valueType

		if column.IsCurrentTimeStamp() {
			tpl.CurTimeStamp = append(tpl.CurTimeStamp, fieldName)
		}

		if column.IsOnUpdate() {
			tpl.OnUpdateTimeStamp = append(tpl.OnUpdateTimeStamp, fieldName)
			tpl.OnUpdateTimeStampStr = append(tpl.OnUpdateTimeStampStr,
				column.ColumnName)
		}

		doc = column.FilterComment()
		if doc != "" {
			field.Doc = append(field.Doc, "// "+doc)
		}

		primary := ""
		if column.IsPri() {
			primary = ";primary_key"
			tpl.EntitySign = field.Name
		}

		if !nullable {
			colDefault = column.GetDefCol()
		}

		field.Tag = fmt.Sprintf("`gorm:\"column:%s%s%s\"`", column.ColumnName, primary, colDefault)

		delField = column.CheckDelField()
		if delField != "" {
			tpl.DelField = delField
		}

		field.Field = field.Name + " " + field.Type
		structInfo.Fields = append(structInfo.Fields, field)

		colDefault = ""
		nullable = false
	}

	structInfo.StructName = tpl.StructName

	tpl.StructInfo = structInfo

	edf.data = tpl
}

// execute implements DomainFile.
func (edf *entityDomainFile) Execute() string {
	content, err := edf.tpl.Execute(edf.name, edf.template, edf.data)
	if err != nil {
		edf.logger.Panicf(err.Error())
	}

	return content
}

// getSavePath implements DomainFile.
func (edf *entityDomainFile) GetSavePath() string {
	return edf.withEntityTarget + edf.tableName + DomainFileExt
}

func (edf *entityDomainFile) GetName() string {
	return edf.name
}

func (edf *entityDomainFile) GetInjectInfo() *InjectInfo {
	return NewInjectInfo()
}
