package domainfile

import (
	"errors"
	"path/filepath"
	"strings"

	"code.jshyjdtech.com/godev/hykit/log"
	"code.jshyjdtech.com/godev/hykit/pkg"
	filedir "code.jshyjdtech.com/godev/hykit/pkg/file-dir"
	"code.jshyjdtech.com/godev/hykit/pkg/templates"
	"github.com/spf13/viper"
)

type repoDomainFile struct {
	withRepoTarget string

	withDisableRepo bool

	name string

	shareInfo *ShareInfo

	template string

	data *repoTpl

	logger log.Logger

	tpl templates.Tpl

	tableName string
}

type RepoDomainFileOption func(*repoDomainFile)

func NewRepoDomainFile(options ...RepoDomainFileOption) DomainFile {
	e := &repoDomainFile{}

	for _, option := range options {
		option(e)
	}

	e.name = "repo"

	e.template = repoTemplate

	return e
}

func WithRepoDomainFileLogger(logger log.Logger) RepoDomainFileOption {
	return func(e *repoDomainFile) {
		e.logger = logger
	}
}

func WithRepoDomainFileTpl(tpl templates.Tpl) RepoDomainFileOption {
	return func(e *repoDomainFile) {
		e.tpl = tpl
	}
}

// Disabled implements DomainFile.
func (rdf *repoDomainFile) Disabled() bool {
	return rdf.withDisableRepo
}

// BindInput implements DomainFile.
//nolint:dupl
func (rdf *repoDomainFile) BindInput(v *viper.Viper) error {
	rdf.withDisableRepo = v.GetBool("disable_repo")
	if !rdf.withDisableRepo {
		rdf.withRepoTarget = v.GetString("repo_target")
		if rdf.withRepoTarget == "" {
			rdf.withRepoTarget = "internal" + string(filepath.Separator) +
				"infra" + string(filepath.Separator) + "repo"
		} else {
			rdf.withRepoTarget = strings.TrimLeft(rdf.withRepoTarget, ".") +
				string(filepath.Separator)
			rdf.withRepoTarget = strings.Trim(rdf.withRepoTarget, "/")
		}

		// check repo dir
		existsRepo, err := filedir.IsExistsDir(rdf.withRepoTarget)
		if err != nil {
			return err
		}

		if !existsRepo {
			return errors.New("repo dir not exists")
		}

		rdf.withRepoTarget += string(filepath.Separator)

		rdf.logger.Debugf("withRepoTarget %s", rdf.withRepoTarget)
	}

	return nil
}

// ParseCloumns implements DomainFile.
func (rdf *repoDomainFile) ParseCloumns(cs Columns, info *ShareInfo) {
	rdf.shareInfo = info

	repoTpl := newRepoTpl(info.CamelStruct)

	if cs.Len() == 0 {
		return
	}

	repoTpl.TableName = info.DbConf.Table
	rdf.tableName = info.DbConf.Table

	repoTpl.Imports = append(
		repoTpl.Imports, pkg.Import{Path: "context"},
		pkg.Import{Path: "code.jshyjdtech.com/godev/hykit/log"},
		pkg.Import{Path: filedir.GetGoProPath() + pkg.DirPathToImportPath(info.WithEntityTarget)},
		pkg.Import{Path: filedir.GetGoProPath() + pkg.DirPathToImportPath(info.WithDaoTarget)})

	for i := range cs {
		column := (&cs[i])

		repoTpl.DelField = column.CheckDelField()
	}

	rdf.data = repoTpl
}

// Execute implements DomainFile.
func (rdf *repoDomainFile) Execute() string {
	content, err := rdf.tpl.Execute(rdf.name, rdf.template, rdf.data)
	if err != nil {
		rdf.logger.Panicf(err.Error())
	}

	return content
}

// GetSavePath implements DomainFile.
func (rdf *repoDomainFile) GetSavePath() string {
	return rdf.withRepoTarget + rdf.tableName + DomainFileExt
}

func (rdf *repoDomainFile) GetName() string {
	return rdf.name
}

func (rdf *repoDomainFile) GetInjectInfo() *InjectInfo {
	injectInfo := NewInjectInfo()

	field := pkg.Field{}
	field.Name = rdf.shareInfo.CamelStruct + "Repo"
	field.Type = " repo." + rdf.shareInfo.CamelStruct + "Repo"
	field.Field = field.Name + " " + field.Type
	injectInfo.Fields = append(injectInfo.Fields, field)

	injectInfo.InfraSetArgs = append(injectInfo.InfraSetArgs,
		"provide"+rdf.shareInfo.CamelStruct+"Repo")

	path := filedir.GetGoProPath() + pkg.DirPathToImportPath(rdf.shareInfo.WithRepoTarget)
	provideRepoFunc := NewProvideRepoFunc(rdf.shareInfo.CamelStruct, path)
	injectInfo.ProvideRepoFuns = append(injectInfo.ProvideRepoFuns, provideRepoFunc)

	return injectInfo
}
