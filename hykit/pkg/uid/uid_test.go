package uid

import (
	"os"
	"testing"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
)

var logger log.Logger
var memConfig config.Config

func TestMain(m *testing.M) {

	logger = log.NewLogger(
		log.WithDebug(true),
	)
	memConfig = config.NewMemConfig()

	memConfig.Set("debug", true)
	logger.Infof("test")

	code := m.Run()

	os.Exit(code)

}

func TestNewUIDRepo(t *testing.T) {

	//memConfig.Set("uid_machine_id", 1234)
	//memConfig.Set("uid_start_time", "20160101")

	uidOptions := UIDOptions{}
	uidRepo := NewUIDRepo(uidOptions.WithConf(memConfig))

	logger.Infof("test")
	logger.Infof("trade_id[%s]", uidRepo.TradeID())
	logger.Infof("trade_id[%s]", uidRepo.TradeID())
	logger.Infof("trade_id[%s]", uidRepo.TradeID())
	logger.Infof("trade_id[%s]", uidRepo.TradeID())
	logger.Infof("trade_id[%s]", uidRepo.TradeID())

}
