package factory

import (
	"go/token"
	"os"
	"reflect"
	"strings"
	"testing"

	"code.jshyjdtech.com/godev/hykit/log"
	"code.jshyjdtech.com/godev/hykit/pkg/templates"
	"github.com/dave/dst"
	"github.com/stretchr/testify/assert"
)

const (
	testStructName = "Test"
)

var resultExpectd = `package example

import (
	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	"code.jshyjdtech.com/godev/hykit/pkg"
)

var (
	var1 = []string{"var1"} //nolint:unused,varcheck,deadcode
)

//nolint:unused,structcheck,maligned
type Test struct {
	c int8

	i bool

	g byte

	d int16

	a int32

	q rune

	f float32

	p complex64

	m map[string]interface{}

	o uint

	n func(interface{})

	b int64

	r uintptr

	e string

	hh []interface{}

	h []int

	pkg.Fields

	u [3]string

	pkg.Field

	logger log.Logger

	conf config.Config
}

type TestOption func(*Test)

func NewTest(options ...TestOption) *Test {
	t := &Test{}

	for _, option := range options {
		option(t)
	}

	if t.m == nil {
		t.m = make(map[string]interface{}, 0)
	}

	if t.hh == nil {
		t.hh = make([]interface{}, 0)
	}

	if t.h == nil {
		t.h = make([]int, 0)
	}

	return t
}

//nolint:unused,structcheck,maligned
type Test1 struct {
	a int
}
`

func TestMain(m *testing.M) {
	code := m.Run()

	os.Exit(code)
}

func TestEsimFactory_ExtendFieldAndSortField(t *testing.T) {
	logger := log.NewLogger(log.WithDebug(true))
	esimfactory := NewEsimFactory(
		WithEsimFactoryLogger(logger),
	)

	esimfactory.structDir = "./example"
	esimfactory.StructName = testStructName
	esimfactory.withOption = true
	esimfactory.withGenLoggerOption = true
	esimfactory.withGenConfOption = true
	esimfactory.WithNew = true
	esimfactory.withStar = true
	esimfactory.withPrint = true
	esimfactory.withSort = true

	esimfactory.UpStructName = templates.FirstToUpper(testStructName)
	esimfactory.ShortenStructName = templates.Shorten(testStructName)
	esimfactory.LowerStructName = strings.ToLower(testStructName)

	ps := esimfactory.loadPackages()

	found := esimfactory.findStruct(ps)
	assert.True(t, found)

	found = esimfactory.checkNewStruct(ps)
	assert.False(t, found)

	esimfactory.withSort = true
	esimfactory.sortField()

	optionDecl := esimfactory.constructOptionTypeFunc()
	esimfactory.newDecls = append(esimfactory.newDecls, optionDecl)

	funcDecl := esimfactory.constructNew()
	esimfactory.newDecls = append(esimfactory.newDecls, funcDecl)

	esimfactory.extendFields()
	esimfactory.constructDecls()
	assert.Equal(t, resultExpectd, esimfactory.newContext())
}

func TestEsimFactory_getNewFuncTypeReturn(t *testing.T) {
	tests := []struct {
		name          string
		structName    string
		withPool      bool
		withStar      bool
		withInterface bool
		InterName     string
		want          interface{}
	}{
		{"normal", "Test", false, false, false, "",
			dst.NewIdent("Test")},
		{"with pool", "Test", true,
			false, false, "", &dst.StarExpr{
				X: dst.NewIdent("Test")}},
		{"with star", "Test", false,
			true, false, "", &dst.StarExpr{
				X: dst.NewIdent("Test")}},
		{"with interface", "", false,
			false, true, "Test", dst.NewIdent("Test")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ef := NewEsimFactory()
			ef.withStar = tt.withStar
			ef.withPool = tt.withPool
			ef.StructName = tt.structName
			ef.withImpIface = tt.InterName
			got := ef.getNewFuncTypeReturn()
			assert.True(t, reflect.DeepEqual(got.List[0].Type, tt.want))
		})
	}
}

func TestEsimFactory_getStructInstan(t *testing.T) {
	tests := []struct {
		name       string
		structName string
		withPool   bool
		withStar   bool
		want       interface{}
	}{
		{"normal", "Test", false, false,
			&dst.CompositeLit{
				Type: dst.NewIdent("Test"),
			}},
		{"with pool", "Test", true,
			false, &dst.TypeAssertExpr{
				X: &dst.CallExpr{
					Fun: &dst.SelectorExpr{
						X:   dst.NewIdent("testPool"),
						Sel: dst.NewIdent("Get"),
					},
				},
				Type: &dst.StarExpr{
					X: dst.NewIdent("Test"),
				},
			}},
		{"with star", "Test", false,
			true, &dst.UnaryExpr{
				Op: token.AND,
				X: &dst.CompositeLit{
					Type: dst.NewIdent("Test"),
				},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ef := NewEsimFactory()
			ef.withStar = tt.withStar
			ef.withPool = tt.withPool
			ef.ShortenStructName = "t"
			ef.LowerStructName = "test"
			ef.StructName = tt.structName
			ef.UpStructName = "Test"
			got := ef.getStructInstan()

			assert.True(t, reflect.DeepEqual(got.(*dst.AssignStmt).Rhs[0], tt.want))
		})
	}
}
