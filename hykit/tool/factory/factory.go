package factory

import (
	"bytes"
	"errors"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
	"strings"

	"code.jshyjdtech.com/godev/hykit/log"
	filedir "code.jshyjdtech.com/godev/hykit/pkg/file-dir"
	"code.jshyjdtech.com/godev/hykit/pkg/templates"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/gopackages"
	"github.com/davecgh/go-spew/spew"
	"github.com/martinusso/inflect"
	"github.com/spf13/viper"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

var fset = token.NewFileSet()

type EsimFactory struct {
	// 需要被处理的结构体名称.
	StructName string

	// 首字母大写.
	UpStructName string

	LowerStructName string

	ShortenStructName string

	// 结构体绝对路径.
	structDir string

	// 结构体所在文件名
	structFileName string

	// 是否找到结构体.
	found bool

	withPlural bool

	withOption bool

	withGenLoggerOption bool

	withGenConfOption bool

	withPrint bool

	logger log.Logger

	withSort bool

	withImpIface string

	withPool bool

	withStar bool

	WithNew bool

	writer filedir.IfaceWriter

	SpecFieldInitStr string

	typeSpec *dst.TypeSpec

	underType *types.Struct

	structPackage *decorator.Package

	dstFile *dst.File

	structIndex int

	newDecls []dst.Decl
}

type Option func(*EsimFactory)

func NewEsimFactory(options ...Option) *EsimFactory {
	factory := &EsimFactory{}

	for _, option := range options {
		option(factory)
	}

	if factory.logger == nil {
		factory.logger = log.NewLogger()
	}

	if factory.writer == nil {
		factory.writer = filedir.NewEsimWriter()
	}

	factory.newDecls = make([]dst.Decl, 0)

	return factory
}

func WithEsimFactoryLogger(logger log.Logger) Option {
	return func(ef *EsimFactory) {
		ef.logger = logger
	}
}

func WithEsimFactoryWriter(writer filedir.IfaceWriter) Option {
	return func(ef *EsimFactory) {
		ef.writer = writer
	}
}

// getPluralWord 复数形式.
// 如果没找到复数形式，在结尾增加"s".
//nolint:unused
func (ef *EsimFactory) getPluralForm(word string) string {
	newWord := inflect.Pluralize(word)
	if newWord == word || newWord == "" {
		newWord = word + "s"
	}

	return newWord
}

func (ef *EsimFactory) Run(v *viper.Viper) error {
	err := ef.bindInput(v)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	ps := ef.loadPackages()
	if !ef.findStruct(ps) {
		ef.logger.Panicf("Not found struct %s", ef.StructName)
	}

	if ef.withSort {
		ef.sortField()
	}

	// if ef.withPool {
	//	decl := ef.constructVarPool()
	//	ef.dstFile.Decls = append(ef.dstFile.Decls, decl)
	// }

	if ef.withOption {
		decl := ef.constructOptionTypeFunc()
		ef.newDecls = append(ef.newDecls, decl)
	}

	if ef.WithNew {
		if ef.checkNewStruct(ps) {
			ef.logger.Panicf("%s is exists.", "New"+ef.UpStructName)
		}

		decl := ef.constructNew()
		ef.newDecls = append(ef.newDecls, decl)
	}

	// if ef.withPool {
	//	decl := ef.constructRelease()
	//	ef.dstFile.Decls = append(ef.dstFile.Decls, decl)
	// }

	ef.extendFields()

	ef.constructDecls()

	if ef.withPrint {
		ef.printResult()
	} else {
		err = filedir.EsimBackUpFile(ef.structDir +
			string(filepath.Separator) + ef.structFileName)
		if err != nil {
			ef.logger.Warnf("backup err %s:%s", ef.structDir+
				string(filepath.Separator)+ef.structFileName,
				err.Error())
		}

		originContent := ef.newContext()
		res, err := imports.Process("", []byte(originContent), nil)
		if err != nil {
			ef.logger.Panicf("%s : %s", err.Error(), originContent)
		}

		err = filedir.EsimWrite(ef.structDir+
			string(filepath.Separator)+ef.structFileName,
			string(res))
		if err != nil {
			ef.logger.Panicf(err.Error())
		}
	}

	return nil
}

//nolint:staticcheck
func (ef *EsimFactory) loadPackages() []*decorator.Package {
	var conf loader.Config
	conf.TypeChecker.Sizes = types.SizesFor(build.Default.Compiler, build.Default.GOARCH)
	conf.Fset = fset

	pConfig := &packages.Config{}
	pConfig.Fset = fset
	// pConfig.Mode = packages.NeedSyntax | packages.NeedName | packages.NeedFiles |
	// packages.NeedCompiledGoFiles | packages.NeedTypesInfo | packages.NeedDeps |
	// packages.NeedTypes | packages.NeedTypesSizes
	pConfig.Mode = packages.LoadAllSyntax
	pConfig.Dir = ef.structDir
	absDir, err := filepath.Abs(ef.structDir)
	if err != nil {
		ef.logger.Warnf(err.Error())
	}

	pConfig.ParseFile = func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
		var f *ast.File
		const mode = parser.ParseComments | parser.AllErrors
		f, err = parser.ParseFile(fset, filename, src, mode)
		if f.Scope.Lookup(ef.StructName) != nil && absDir == filepath.Dir(filename) {
			ef.structFileName = filepath.Base(filename)
		}
		return f, err
	}
	ps, err := decorator.Load(pConfig)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	return ps
}

// 查找要被处理的结构体.
func (ef *EsimFactory) findStruct(ps []*decorator.Package) bool {
	for _, p := range ps {
		for _, syntax := range p.Syntax {
			for k, decl := range syntax.Decls {
				if genDecl, ok := decl.(*dst.GenDecl); ok {
					if genDecl.Tok == token.TYPE {
						for _, spec := range genDecl.Specs {
							if typeSpec, ok := spec.(*dst.TypeSpec); ok {
								for _, def := range p.TypesInfo.Defs {
									if def == nil {
										continue
									}

									if _, ok := def.(*types.TypeName); !ok {
										continue
									}

									typ, ok := def.Type().(*types.Named)
									if !ok {
										continue
									}

									underType, ok := typ.Underlying().(*types.Struct)
									if !ok {
										continue
									}

									if typ.Obj().Name() == ef.StructName && !ef.found {
										ef.found = true
										ef.typeSpec = typeSpec
										ef.underType = underType
										ef.structPackage = p
										ef.dstFile = syntax
										ef.structIndex = k
										// ef.structFileName =
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return ef.found
}

func (ef *EsimFactory) constructOptionTypeFunc() *dst.GenDecl {
	if !ef.withOption {
		return &dst.GenDecl{}
	}

	genDecl := &dst.GenDecl{
		Tok: token.TYPE,
		Specs: []dst.Spec{
			&dst.TypeSpec{
				Name: dst.NewIdent(ef.StructName + "Option"),
				Type: &dst.FuncType{
					Func:   true,
					Params: ef.constructOptionTypeFuncParam(),
				},
			},
		},
		Decs: dst.GenDeclDecorations{
			NodeDecs: dst.NodeDecs{
				Before: dst.EmptyLine,
			},
		},
	}

	return genDecl
}

func (ef *EsimFactory) constructOptionTypeFuncParam() *dst.FieldList {
	fieldList := &dst.FieldList{}
	if ef.withPool || ef.withStar {
		fieldList.List = []*dst.Field{
			{
				Type: &dst.StarExpr{
					X: dst.NewIdent(ef.StructName),
				},
			},
		}
	} else {
		fieldList.List = []*dst.Field{
			{
				Type: dst.NewIdent(ef.StructName),
			},
		}
	}

	return fieldList
}

func (ef *EsimFactory) checkNewStruct(ps []*decorator.Package) bool {
	for _, p := range ps {
		for _, syntax := range p.Syntax {
			for _, decl := range syntax.Decls {
				if genDecl, ok := decl.(*dst.FuncDecl); ok {
					if genDecl.Name.String() == "New"+ef.UpStructName {
						return true
					}
				}
			}
		}
	}

	return false
}

func (ef *EsimFactory) constructNew() *dst.FuncDecl {
	params := &dst.FieldList{}
	if ef.withOption {
		params = &dst.FieldList{
			Opening: true,
			List: []*dst.Field{
				{
					Names: []*dst.Ident{dst.NewIdent("options")},
					Type: &dst.Ellipsis{
						Elt: dst.NewIdent(ef.UpStructName + "Option"),
					},
				},
			},
		}
	}

	stmts := make([]dst.Stmt, 0)
	stmts = append(stmts, ef.getStructInstan(), ef.getOptionBody())
	stmts = append(stmts, ef.getSpecialFieldStmt()...)

	stmts = append(stmts,
		&dst.ReturnStmt{
			Results: []dst.Expr{
				dst.NewIdent(ef.ShortenStructName),
			},
			Decs: dst.ReturnStmtDecorations{
				NodeDecs: dst.NodeDecs{
					Before: dst.EmptyLine,
				},
			},
		})

	funcDecl := &dst.FuncDecl{
		Name: dst.NewIdent("New" + ef.UpStructName),
		Type: &dst.FuncType{
			Func:    true,
			Params:  params,
			Results: ef.getNewFuncTypeReturn(),
		},
		Body: &dst.BlockStmt{
			List: stmts,
		},
	}

	return funcDecl
}

//nolint:unused
func (ef *EsimFactory) constructRelease() *dst.FuncDecl {
	funcDecl := &dst.FuncDecl{
		Recv: &dst.FieldList{
			Opening: true,
			List: []*dst.Field{
				{
					Names: []*dst.Ident{
						dst.NewIdent("t"),
					},
					Type: &dst.StarExpr{
						X: dst.NewIdent("Test"),
					},
				},
			},
		},
		Name: dst.NewIdent("Reset"),
		Type: &dst.FuncType{
			Func: true,
			Params: &dst.FieldList{
				Opening: true,
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{

				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun: &dst.SelectorExpr{
							X:   dst.NewIdent("testPool"),
							Sel: dst.NewIdent("Put"),
						},
						Args: []dst.Expr{
							dst.NewIdent("t"),
						},
					},
				},
			},
		},
	}

	_ = funcDecl

	return funcDecl
}

// t := Test{}.
func (ef *EsimFactory) getStructInstan() dst.Stmt {
	if ef.withPool {
		return &dst.AssignStmt{
			Lhs: []dst.Expr{
				dst.NewIdent(ef.ShortenStructName),
			},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.TypeAssertExpr{
					X: &dst.CallExpr{
						Fun: &dst.SelectorExpr{
							X:   dst.NewIdent(ef.LowerStructName + "Pool"),
							Sel: dst.NewIdent("Get"),
						},
					},
					Type: &dst.StarExpr{
						X: dst.NewIdent(ef.UpStructName),
					},
				},
			},
		}
	} else if ef.withStar {
		return &dst.AssignStmt{
			Lhs: []dst.Expr{
				dst.NewIdent(ef.ShortenStructName),
			},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.UnaryExpr{
					Op: token.AND,
					X: &dst.CompositeLit{
						Type: dst.NewIdent(ef.UpStructName),
					},
				},
			},
		}
	} else {
		return &dst.AssignStmt{
			Lhs: []dst.Expr{
				dst.NewIdent(ef.ShortenStructName),
			},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.CompositeLit{
					Type: dst.NewIdent(ef.UpStructName),
				},
			},
		}
	}
}

func (ef *EsimFactory) getOptionBody() dst.Stmt {
	return &dst.RangeStmt{
		Key:   dst.NewIdent("_"),
		Value: dst.NewIdent("option"),
		Tok:   token.DEFINE,
		X:     dst.NewIdent("options"),
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun: dst.NewIdent("option"),
						Args: []dst.Expr{
							dst.NewIdent(ef.ShortenStructName),
						},
					},
				},
			},
		},
		Decs: dst.RangeStmtDecorations{
			NodeDecs: dst.NodeDecs{
				Before: dst.EmptyLine,
			},
		},
	}
}

// map,slice are special field.
func (ef *EsimFactory) getSpecialFieldStmt() []dst.Stmt {
	stmts := make([]dst.Stmt, 0)
	for _, field := range ef.typeSpec.Type.(*dst.StructType).Fields.List {
		field.Decs.After = dst.EmptyLine
		switch _typ := field.Type.(type) {
		case *dst.ArrayType:
			if len(field.Names) != 0 && _typ.Len == nil {
				cloned := dst.Clone(_typ).(*dst.ArrayType)
				stmts = append(stmts, ef.constructSpecialFieldStmt(ef.ShortenStructName,
					field.Names[0].String(), cloned))
			}
		case *dst.MapType:
			if len(field.Names) != 0 {
				cloned := dst.Clone(_typ).(*dst.MapType)
				stmts = append(stmts, ef.constructSpecialFieldStmt(ef.ShortenStructName,
					field.Names[0].String(), cloned))
			}
		}
	}

	return stmts
}

//nolint:unparam,unused
func (ef *EsimFactory) getReleaseFieldStmt() []dst.Stmt {
	// numField := ef.underType.NumFields()
	for _, field := range ef.typeSpec.Type.(*dst.StructType).Fields.List {
		field.Decs.After = dst.EmptyLine
		// spew.Dump(ef.underType.Field(k).Type().String())

		switch _t := field.Type.(type) {
		case *dst.ArrayType:
		case *dst.MapType:
		default:
			spew.Dump(_t)
		}
	}

	return nil
}

func (ef *EsimFactory) TypeToInit(ident *dst.Ident) string {
	var initStr string

	switch ident.Name {
	case "string":
		initStr = "\"\""
	case "int", "int64", "int8", "int16", "int32":
		initStr = "0"
	case "uint", "uint64", "uint8", "uint16", "uint32":
		initStr = "0"
	case "bool":
		initStr = "false"
	case "float32", "float64":
		initStr = "0.00"
	case "complex64", "complex128":
		initStr = "0+0i"
		// case reflect.Interface:
		//	initStr = "nil"
		// case reflect.Uintptr:
		//	initStr = "0"
		// case reflect.Invalid, reflect.Func, reflect.Chan, reflect.Ptr, reflect.UnsafePointer:
		//	initStr = "nil"
		// case reflect.Slice:
		//	initStr = "nil"
		// case reflect.Map:
		//	initStr = "nil"
		// case reflect.Array:
		//	initStr = "nil"
	}

	return initStr
}

func (ef *EsimFactory) getNewFuncTypeReturn() *dst.FieldList {
	fieldList := &dst.FieldList{}
	if ef.withImpIface != "" {
		fieldList.List = []*dst.Field{
			{
				Type: dst.NewIdent(ef.withImpIface),
			},
		}
	} else if ef.withPool || ef.withStar {
		fieldList.List = []*dst.Field{
			{
				Type: &dst.StarExpr{
					X: dst.NewIdent(ef.StructName),
				},
			},
		}
	} else {
		fieldList.List = []*dst.Field{
			{
				Type: dst.NewIdent(ef.StructName),
			},
		}
	}

	return fieldList
}

//nolint:unused
func (ef *EsimFactory) constructVarPool() *dst.GenDecl {
	valueSpec := &dst.ValueSpec{
		Names: []*dst.Ident{
			{
				Name: ef.LowerStructName + "Pool",
			},
		},
		Values: []dst.Expr{
			&dst.CompositeLit{
				Type: &dst.SelectorExpr{
					X:   &dst.Ident{Name: "sync"},
					Sel: &dst.Ident{Name: "Pool"},
					Decs: dst.SelectorExprDecorations{
						NodeDecs: dst.NodeDecs{
							End: dst.Decorations{"\n"},
						},
					},
				},
				Elts: []dst.Expr{
					&dst.KeyValueExpr{
						Key: &dst.Ident{Name: "New"},
						Value: &dst.FuncLit{
							Type: &dst.FuncType{
								Func: true,
								Params: &dst.FieldList{
									Opening: true,
									Closing: true,
								},
								Results: &dst.FieldList{
									List: []*dst.Field{
										{
											Type: &dst.InterfaceType{
												Methods: &dst.FieldList{
													Opening: true,
													Closing: true,
												},
											},
										},
									},
								},
							},
							Body: &dst.BlockStmt{
								List: []dst.Stmt{
									&dst.ReturnStmt{
										Results: []dst.Expr{
											&dst.UnaryExpr{
												Op: token.AND,
												X: &dst.CompositeLit{
													Type: &dst.Ident{Name: ef.StructName},
												},
											},
										},
										Decs: dst.ReturnStmtDecorations{
											NodeDecs: dst.NodeDecs{
												Before: dst.NewLine,
											},
										},
									},
								},
							},
						},
						Decs: dst.KeyValueExprDecorations{
							NodeDecs: dst.NodeDecs{
								Before: dst.NewLine,
							},
						},
					},
				},
			},
		},
	}

	genDecl := &dst.GenDecl{
		Tok:    token.VAR,
		Lparen: true,
		Specs: []dst.Spec{
			valueSpec,
		},
	}

	return genDecl
}

func (ef *EsimFactory) constructSpecialFieldStmt(structName, field string, expr dst.Expr) dst.Stmt {
	stmt := &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X: &dst.SelectorExpr{
				X:   dst.NewIdent(structName),
				Sel: dst.NewIdent(field),
			},
			Op: token.EQL,
			Y:  dst.NewIdent("nil"),
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.AssignStmt{
					Lhs: []dst.Expr{
						&dst.SelectorExpr{
							X:   dst.NewIdent(structName),
							Sel: dst.NewIdent(field),
						},
					},
					Tok: token.ASSIGN,
					Rhs: []dst.Expr{
						&dst.CallExpr{
							Fun: dst.NewIdent("make"),
							Args: []dst.Expr{
								expr,
								&dst.BasicLit{
									Kind:  token.INT,
									Value: "0",
								},
							},
						},
					},
				},
			},
		},
		Decs: dst.IfStmtDecorations{
			NodeDecs: dst.NodeDecs{
				Before: dst.EmptyLine,
			},
		},
	}

	return stmt
}

func (ef *EsimFactory) newContext() string {
	r := decorator.NewRestorerWithImports("root", gopackages.New(ef.structDir))

	var buf bytes.Buffer
	if err := r.Fprint(&buf, ef.dstFile); err != nil {
		panic(err)
	}

	return buf.String()
}

func (ef *EsimFactory) constructDecls() {
	if ef.structIndex == len(ef.dstFile.Decls) {
		ef.dstFile.Decls = append(ef.dstFile.Decls, ef.newDecls...)
	} else {
		head := ef.dstFile.Decls[:ef.structIndex+1]
		tail := ef.dstFile.Decls[ef.structIndex+1:]
		ef.dstFile.Decls = head
		ef.dstFile.Decls = append(ef.dstFile.Decls, ef.newDecls...)
		ef.dstFile.Decls = append(ef.dstFile.Decls, tail...)
	}
}

// printResult 在终端打印内容.
func (ef *EsimFactory) printResult() {
	println(ef.newContext())
}

func (ef *EsimFactory) bindInput(v *viper.Viper) error {
	sname := v.GetString("sname")
	if sname == "" {
		return errors.New("sname is empty")
	}
	ef.StructName = sname
	ef.UpStructName = templates.FirstToUpper(sname)
	ef.ShortenStructName = templates.Shorten(sname)
	ef.LowerStructName = strings.ToLower(sname)

	sdir := v.GetString("sdir")
	if sdir == "" {
		sdir = "."
	}

	dir, err := filepath.Abs(sdir)
	if err != nil {
		return err
	}
	ef.structDir = strings.TrimRight(dir, "/")

	plural := v.GetBool("plural")
	if plural {
		ef.withPlural = true
	}

	ef.withOption = v.GetBool("option")

	ef.withGenConfOption = v.GetBool("oc")

	ef.withGenLoggerOption = v.GetBool("ol")

	ef.withSort = v.GetBool("sort")

	ef.withImpIface = v.GetString("imp_iface")

	ef.withPool = v.GetBool("pool")

	ef.withStar = v.GetBool("star")

	ef.withPrint = v.GetBool("print")

	ef.WithNew = v.GetBool("new")

	return nil
}

// 为结构体扩展 logger 和 conf 属性.
// 必须在构造后.
func (ef *EsimFactory) extendFields() {
	if ef.withOption {
		if ef.withGenLoggerOption {
			ef.extendField(newLoggerFieldInfo())
		}

		if ef.withGenConfOption {
			ef.extendField(newConfigFieldInfo())
		}
	}
}

type FieldSizes []FieldSize

type FieldSize struct {
	Size int64

	Field *dst.Field

	Vars *types.Var
}

func (fs FieldSizes) Len() int { return len(fs) }

func (fs FieldSizes) Less(i, j int) bool {
	return fs[i].Size < fs[j].Size
}

func (fs FieldSizes) Swap(i, j int) { fs[i], fs[j] = fs[j], fs[i] }

func (fs FieldSizes) getFields() []*dst.Field {
	dstFields := make([]*dst.Field, 0)
	for _, f := range fs {
		dstFields = append(dstFields, f.Field)
	}

	return dstFields
}

// sortField 以字节大小排序.
func (ef *EsimFactory) sortField() {
	fs := make(FieldSizes, 0)
	for k, field := range ef.typeSpec.Type.(*dst.StructType).Fields.List {
		field.Decs.After = dst.EmptyLine
		fs = append(fs, FieldSize{
			Size:  ef.structPackage.TypesSizes.Sizeof(ef.underType.Field(k).Type()),
			Field: field,
		},
		)
	}
	sort.Sort(fs)
	ef.typeSpec.Type.(*dst.StructType).Fields.List = fs.getFields()
}

func (ef *EsimFactory) extendField(fieldInfo extendFieldInfo) {
	fields := ef.typeSpec.Type.(*dst.StructType).Fields.List
	var fieldExists bool
	for _, field := range fields {
		if len(field.Names) != 0 {
			if field.Names[0].Name == fieldInfo.name {
				fieldExists = true
			}
		}
	}

	if !fieldExists {
		fields = append(fields,
			&dst.Field{
				Names: []*dst.Ident{
					dst.NewIdent(fieldInfo.name),
				},
				Type: &dst.Ident{Name: fieldInfo.typeName, Path: fieldInfo.typePath},
				Decs: dst.FieldDecorations{
					NodeDecs: dst.NodeDecs{
						Before: dst.EmptyLine,
					},
				},
			},
		)
		ef.typeSpec.Type.(*dst.StructType).Fields.List = fields
	}
}
