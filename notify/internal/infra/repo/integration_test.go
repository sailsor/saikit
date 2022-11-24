package repo

import (
	"os"
	"testing"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/mysql"
)

var mysqlClient *mysql.Client

func TestMain(m *testing.M) {
	clientOptions := mysql.ClientOptions{}

	options := config.ViperConfOptions{}
	confFile := "../../../conf/dev.yaml"
	file := []string{confFile}
	conf := config.NewViperConfig(options.WithConfigType("yaml"),
		options.WithConfFile(file))

	mysqlClient = mysql.NewClient(clientOptions.WithConf(conf))

	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {

}

func tearDown() {
	mysqlClient.Close()
}
