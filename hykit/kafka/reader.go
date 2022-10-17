package kafka

import (
	"context"
	"errors"
	"time"

	"github.com/segmentio/kafka-go"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
)

const (
	FirstOffset = kafka.FirstOffset
	LastOffset  = kafka.LastOffset
)

type Message = kafka.Message

type Reader struct {
	reader *kafka.Reader

	brokers []string

	groupId string

	topicName string

	maxNum int //最多消息数目

	maxWait int //最大等待时间

	logger log.Logger

	conf config.Config
}

type ReaderOption func(*Reader)

func WithReaderLogger(logger log.Logger) ReaderOption {
	return func(se *Reader) {
		se.logger = logger
	}
}

func WithReaderConf(conf config.Config) ReaderOption {
	return func(se *Reader) {
		se.conf = conf
	}
}

func WithReaderTopic(topicName string) ReaderOption {
	return func(r *Reader) {
		r.topicName = topicName
	}
}

func WithReaderGroupID(groupID string) ReaderOption {
	return func(r *Reader) {
		r.groupId = groupID
	}
}

func WithReaderMaxWait(maxWait int) ReaderOption {
	return func(r *Reader) {
		r.maxWait = maxWait
	}
}

func NewReader(options ...ReaderOption) *Reader {
	r := &Reader{}

	for _, option := range options {
		option(r)
	}

	if r.conf == nil {
		r.conf = config.NewNullConfig()
	}

	if r.logger == nil {
		r.logger = log.NewLogger()
	}

	// 服务器
	r.brokers = r.conf.GetStringSlice("kafka_brokers")
	if r.brokers == nil {
		r.logger.Panicf("kafka_brokers is null! please confirm")
	}

	// topicName
	if r.topicName == "" {
		r.topicName = r.conf.GetString("kafka_topic")
		if r.topicName == "" {
			r.logger.Panicf("kafka_topic is null! please confirm")
		}
	}

	//等待消息时间 单位ms
	if r.maxWait == 0 {
		r.maxWait = r.conf.GetInt("kafka_max_wait")
		if r.maxWait == 0 {
			r.maxWait = 500
		}
	}

	//批量读取消息数目
	if r.maxNum == 0 {
		r.maxNum = r.conf.GetInt("kafka_max_num")
	}

	r.logger.Infof("brokers[%v] groupId[%s] topic[%s]", r.brokers, r.groupId, r.topicName)
	r.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  r.brokers,
		GroupID:  r.groupId,
		Topic:    r.topicName,
		MinBytes: 1,    // 1byte 收到后返回
		MaxBytes: 10e6, // 10MB
		MaxWait:  time.Duration(r.maxWait) * time.Millisecond,
		//Logger:   r.logger,
	})

	return r
}

func (r *Reader) KafkaReader() *kafka.Reader {
	return r.reader
}

func (r *Reader) FetchMessage(ctx context.Context) (*Message, error) {
	m, err := r.reader.FetchMessage(ctx)
	if err != nil {
		return nil, err
	}
	r.logger.Infoc(ctx, "message at topic:%s partition:%d offset:%d key:%s", m.Topic, m.Partition, m.Offset, string(m.Key))
	return &m, nil
}

func (r *Reader) CommitMessage(ctx context.Context, m *Message) error {
	if r.groupId == "" {
		r.logger.Errorc(ctx, "CommitMessage must in consumer group")
		return errors.New("CommitMessage must in consumer group")
	}

	return r.reader.CommitMessages(ctx, *m)
}

func (r *Reader) RollbackMessage(ctx context.Context, m *Message) error {
	if r.groupId != "" {
		r.logger.Errorc(ctx, "RollbackMessage must not in consumer group")
		return errors.New("RollbackMessage must not in consumer group")
	}

	return r.reader.SetOffset(m.Offset)
}

func (r *Reader) SetOffset(ctx context.Context, offset int64) error {
	if r.groupId != "" {
		r.logger.Errorc(ctx, "SetOffset must not in consumer group")
		return errors.New("SetOffset must not in consumer group")
	}
	return r.reader.SetOffset(offset)
}

func (r *Reader) SetOffSetAt(ctx context.Context, tm time.Time) error {
	if r.groupId != "" {
		r.logger.Errorc(ctx, "SetOffSetAt must not in consumer group")
		return errors.New("SetOffSetAt must not in consumer group")
	}
	return r.reader.SetOffsetAt(ctx, tm)
}
