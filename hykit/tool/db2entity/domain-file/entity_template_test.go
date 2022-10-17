package domainfile

import (
	"bytes"
	"testing"
	"text/template"

	"code.jshyjdtech.com/godev/hykit/pkg"
	"code.jshyjdtech.com/godev/hykit/pkg/templates"
	"github.com/stretchr/testify/assert"
)

func TestEntityTemplate(t *testing.T) {
	tmpl, err := template.New("entity_template").Funcs(templates.EsimFuncMap()).
		Parse(entityTemplate)
	assert.Nil(t, err)

	var imports pkg.Imports
	imports = append(imports, pkg.Import{Name: "time", Path: "time"},
		pkg.Import{Name: "sync", Path: "sync"})

	Field1 := pkg.Field{}
	Field1.Name = "id"
	Field1.Field = "id int"
	Field1.Tag = "`json:\"id\"`"

	Field2 := pkg.Field{}
	Field2.Name = "name"
	Field2.Field = "name string"
	Field2.Tag = "`json:\"name\"`"
	Field2.Doc = append(Field2.Doc, "//username \\r\\n is a test")

	var buf bytes.Buffer
	tpl := entityTpl{}
	tpl.StructName = "Entity"
	tpl.CurTimeStamp = append(tpl.CurTimeStamp, "CreateTime1", "CreateTime2")

	tpl.OnUpdateTimeStamp = append(tpl.OnUpdateTimeStamp, "LastUpdateTime")

	tpl.OnUpdateTimeStampStr = append(tpl.OnUpdateTimeStampStr,
		"last_update_time1", "last_update_time2")
	tpl.EntitySign = "id"

	tpl.Imports = imports
	tpl.DelField = delField

	structInfo := templates.NewStructInfo()
	structInfo.StructName = tpl.StructName
	structInfo.Fields = append(structInfo.Fields, Field1, Field2)

	tpl.StructInfo = structInfo

	err = tmpl.Execute(&buf, tpl)
	println(buf.String())
	assert.Nil(t, err)
}
