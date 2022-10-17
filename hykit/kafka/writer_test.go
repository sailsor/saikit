package kafka

import (
	"context"
	"testing"
	"time"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	"github.com/stretchr/testify/assert"
)

var logger log.Logger = log.NewLogger()

const (
	address = "127.0.0.1:9092"

	topicName = "delay-queue-1-minute"

	groupId = "test1"
)

func TestNewWriter(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("kafka_brokers", []string{address})

	ctx := context.Background()
	writer := NewWriter(
		WithWriterLogger(logger),
		WithWriterConf(memConfig))

	err := writer.PublishMessage(ctx, topicName, time.Now().String())
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.Nil(t, err)
	}

	for i := 0; i < 10; i++ {
		err = writer.PublishDelayMessage(ctx, topicName, time.Now().String(), i)
		if err != nil {
			logger.Errorf(err.Error())
		} else {
			assert.Nil(t, err)
		}
	}

}
