package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	kafka "github.com/segmentio/kafka-go"
)

type Writer struct {
	writer *kafka.Writer

	brokers []string

	logger log.Logger

	conf config.Config
}

type WriterOption func(*Writer)

var writeOne sync.Once
var writer *Writer

func NewWriter(options ...WriterOption) *Writer {
	writeOne.Do(func() {
		p := &Writer{}

		for _, option := range options {
			option(p)
		}

		if p.conf == nil {
			p.conf = config.NewNullConfig()
		}

		if p.logger == nil {
			p.logger = log.NewLogger()
		}

		p.brokers = p.conf.GetStringSlice("kafka_brokers")
		if p.brokers == nil {
			p.logger.Panicf("kafka_brokers is null! please confirm")
		}

		/*初始化生产者*/
		p.writer = &kafka.Writer{
			Addr:         kafka.TCP(p.brokers...),
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 100 * time.Millisecond, //100ms 写一次
			BatchSize:    1,                      //有1条就发送
			Logger:       p.logger,
			RequiredAcks: kafka.RequireOne,
		}
		writer = p
	})

	return writer
}

func WithWriterConf(conf config.Config) WriterOption {
	return func(p *Writer) {
		p.conf = conf
	}
}

func WithWriterLogger(logger log.Logger) WriterOption {
	return func(p *Writer) {
		p.logger = logger
	}
}

func (p *Writer) PublishMessage(ctx context.Context, topicName, messageBody string) error {
	if topicName == "" {
		return errors.New("topicName is null")
	}
	msg := kafka.Message{
		Topic: topicName,
		Key:   msgId(),
		Value: []byte(messageBody),
	}
	return p.publish(ctx, msg)
}

func (p *Writer) PublishKeyMessage(ctx context.Context, topicName string, messageKey, messageBody []byte) error {
	if topicName == "" {
		return errors.New("topicName is null")
	}
	msg := kafka.Message{
		Topic: topicName,
		Key:   []byte(messageKey),
		Value: []byte(messageBody),
	}
	return p.publish(ctx, msg)
}

const (
	delayQueue1Minute  = "delay-queue-1-minute"
	delayQueue2Minute  = "delay-queue-2-minute"
	delayQueue4Minute  = "delay-queue-4-minute"
	delayQueue8Minute  = "delay-queue-8-minute"
	delayQueue16Minute = "delay-queue-16-minute"
	delayQueue32Minute = "delay-queue-32-minute"
)

type DelayMessage struct {
	Minute      int
	Topic       string
	MessageBody string
	SendTime    time.Time
}

func (p *Writer) PublishDelayMessage(ctx context.Context, topicName, messageBody string, min int) error {
	if topicName == "" {
		return errors.New("topicName is null")
	}
	if min > 6 {
		min = 6
	}
	if min <= 0 {
		min = 1
	}

	delayQueue := fmt.Sprintf("delay-queue-%d-minute", int(math.Pow(2, float64(min-1))))

	data, err := json.Marshal(&DelayMessage{
		Minute:      min,
		SendTime:    time.Now(),
		Topic:       topicName,
		MessageBody: messageBody,
	})
	if err != nil {
		p.logger.Infoc(ctx, "发送延时消息到[%d][%s] 目的topic[%s]", min, delayQueue, topicName)
		return err
	}

	msg := kafka.Message{
		Topic: delayQueue,
		Key:   msgId(),
		Value: data,
	}
	p.logger.Infoc(ctx, "发送延时消息到[%d][%s] 目的topic[%s]", min, delayQueue, topicName)
	return p.publish(ctx, msg)
}

func (p *Writer) publish(ctx context.Context, msg kafka.Message) error {
	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		p.logger.Errorc(ctx, "WriteMessages失败[%]:[%s]", msg.Topic, err)
		return err
	}
	p.logger.Infoc(ctx, "WriteMessages成功;[%s]", msg.Key)
	return nil
}

func msgId() []byte {
	uid, _ := uuid.NewUUID()
	return []byte(uid.String())
}
