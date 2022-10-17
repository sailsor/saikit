package supermng

import (
	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	"context"
	"testing"
)

func TestGetSuper(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogger()
	clientOptions := ClientOptions{}

	confOps := config.ViperConfOptions{}
	conf := config.NewViperConfig(
		confOps.WithConfPath([]string{"D:\\file\\conf\\dev"}),
		confOps.WithConfigType("yaml"),
		confOps.WithConfFile([]string{"monitoring", "conf", "server"}))

	c := NewClient(
		clientOptions.WithConf(conf),
		clientOptions.WithLogger(logger),
	)

	client := c.GetSuperClient("a")

	_ = client.RestartALL()

	pid, err := client.client.GetPID()
	if err != nil {
		logger.Errorc(ctx, "GetPID err:", err)
		return
	}
	logger.Infof("supervisord pid is %v", pid)
}
