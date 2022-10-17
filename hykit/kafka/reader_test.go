package kafka

import (
	"context"
	"testing"
	"time"

	"code.jshyjdtech.com/godev/hykit/config"
)

func TestNewReader(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("kafka_brokers", []string{address})
	memConfig.Set("kafka_topic", topicName)

	reader := NewReader(
		WithReaderConf(memConfig),
		WithReaderLogger(logger),
		//	WithReaderGroupID(groupId),
		WithReaderTopic("delay-queue-32-minute"),
	)

	ctx := context.Background()

	err := reader.SetOffSetAt(ctx, time.Now().Add(-10*time.Minute))
	if err != nil {
		logger.Errorc(ctx, "SetOffSetAt失败[%s]", err)
	}

	for {
		logger.Infoc(ctx, "begin fetch")
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			logger.Errorc(ctx, "FetchMessage err[%s]", err)
			t.Fatal(err)
		}
		logger.Infoc(ctx, "[%s][%s][%d][%d[", string(m.Key), string(m.Value), m.Partition, m.Offset)
	}

	/*err = reader.RollbackMessage(ctx, m)
	if err != nil {
		logger.Errorc(ctx, "RollbackMessage Err[%s]", err)
		break
	}*/

}
