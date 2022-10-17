package rocketmq

import (
	"time"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	mq_http_sdk "github.com/aliyunmq/mq-http-go-sdk"
)

type Publisher struct {
	client *MQClient

	logger log.Logger

	conf config.Config
}

type PublisherOption func(*Publisher)

func NewPublisher(options ...PublisherOption) *Publisher {

	p := &Publisher{}

	for _, option := range options {
		option(p)
	}

	if p.conf == nil {
		p.conf = config.NewNullConfig()
	}

	if p.logger == nil {
		p.logger = log.NewLogger()
	}

	/*初始化客户端*/
	var cliOpt MQClientOptions
	p.client = NewMQClient(
		cliOpt.WithLogger(p.logger),
		cliOpt.WithConf(p.conf))
	return p
}

func WithPublisherConf(conf config.Config) PublisherOption {
	return func(p *Publisher) {
		p.conf = conf
	}
}

func WithPublisherLogger(logger log.Logger) PublisherOption {
	return func(p *Publisher) {
		p.logger = logger
	}
}

func (p *Publisher) PublishMessage(topicName, messageBody string) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody: messageBody,
	}
	return p.publish(topicName, msg)
}

// with message tag and key
func (p *Publisher) PublishMsgWithTag(topicName, messageBody, messageTag string) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody: messageBody,
		MessageTag:  messageTag,
	}
	return p.publish(topicName, msg)
}

// with message tag
func (p *Publisher) PublishMsgWithKeyTag(topicName, messageBody, messageTag, messageKey string) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody: messageBody,
		MessageTag:  messageTag,
		MessageKey:  messageKey,
	}
	return p.publish(topicName, msg)
}

func (p *Publisher) PublishDelayMessage(topicName, messageBody string, delay time.Duration) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody:      messageBody,
		StartDeliverTime: time.Now().Add(delay).UTC().Unix() * 1000, //值为毫秒级别的Unix时间戳
	}
	return p.publish(topicName, msg)
}

func (p *Publisher) PublishDelayMsgWithTag(topicName, messageBody, messageTag string, delay time.Duration) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody:      messageBody,
		MessageTag:       messageTag,
		StartDeliverTime: time.Now().Add(delay).UTC().Unix() * 1000, //值为毫秒级别的Unix时间戳
	}
	return p.publish(topicName, msg)
}

func (p *Publisher) PublishDelayMsgWithKeyTag(topicName, messageBody, messageTag, messageKey string, delay time.Duration) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody:      messageBody,
		MessageTag:       messageTag,
		MessageKey:       messageKey,
		StartDeliverTime: time.Now().Add(delay).UTC().Unix() * 1000, //值为毫秒级别的Unix时间戳
	}
	return p.publish(topicName, msg)
}

func (p *Publisher) PublishMessageProp(topicName, messageTag, messageBody string, properties map[string]string) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody: messageBody,
		MessageTag:  messageTag, // 消息标签
		Properties:  properties,
	}
	return p.publish(topicName, msg)
}

func (p *Publisher) PublishDelayMessageProt(topicName, messageBody string, properties map[string]string, delay time.Duration) error {
	msg := mq_http_sdk.PublishMessageRequest{
		MessageBody:      messageBody,
		Properties:       properties,
		StartDeliverTime: time.Now().Add(delay).UTC().Unix() * 1000, //值为毫秒级别的Unix时间戳
	}
	return p.publish(topicName, msg)
}

func (p *Publisher) publish(topicName string, msg mq_http_sdk.PublishMessageRequest) error {
	mqProducer := p.client.Producer(topicName)
	resp, err := mqProducer.PublishMessage(msg)
	if err != nil {
		p.logger.Errorf("发布信息失败[%][%s]:[%s]", topicName, err)
		return err
	}
	p.logger.Infof("Publish成功 ---->MessageId:[%s], BodyMD5:[%s];", resp.MessageId, resp.MessageBodyMD5)
	return nil
}
