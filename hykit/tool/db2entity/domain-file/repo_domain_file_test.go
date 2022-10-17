package domainfile

import (
	"os"
	"testing"

	"code.jshyjdtech.com/godev/hykit/log"
	filedir "code.jshyjdtech.com/godev/hykit/pkg/file-dir"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRepoDomainFile(t *testing.T) {
	err := filedir.CreateDir(target)
	assert.Nil(t, err)

	v := viper.New()
	v.Set("repo_target", target)
	v.Set("table", testTable)

	err = testRepoDomainFile.BindInput(v)
	assert.Nil(t, err)

	dbConf := NewDbConfig()
	dbConf.ParseConfig(v, log.NewLogger())

	shareInfo := NewShareInfo()
	shareInfo.CamelStruct = testStructName
	shareInfo.DbConf = dbConf

	testRepoDomainFile.ParseCloumns(Cols, shareInfo)
	content := testRepoDomainFile.Execute()
	assert.NotEmpty(t, content)

	savePath := testRepoDomainFile.GetSavePath()
	assert.Equal(t, "example/test.go", savePath)
	err = os.RemoveAll(target)
	assert.Nil(t, err)
}
