package http

import (
	"fmt"
	"os"
	"testing"

	"code.jshyjdtech.com/godev/hykit/log"
	"github.com/stretchr/testify/assert"
)

var logger log.Logger

const (
	host1 = "http://192.168.3.154:8081/ping"

	host2 = "127.0.0.2"
)

func TestMain(m *testing.M) {

	logger = log.NewLogger()

	code := m.Run()

	os.Exit(code)
}

//nolint:dupl
func TestMulLevelRoundTrip(t *testing.T) {
	clientOptions := ClientOptions{}
	httpClient := NewClient(
		clientOptions.WithLogger(logger),
	)

	cli := httpClient.Client
	cli.SetDebug(true)
	cli.SetLogger(logger)

	resp, err := cli.R().Get("https://www.baidu.com/")
	assert.Nil(t, err)

	fmt.Println(resp.StatusCode())

}
