package rocketmq

import (
	"sync"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	mq_http_sdk "github.com/aliyunmq/mq-http-go-sdk"
)

var poolOnce sync.Once
var onceClient *MQClient

type MQClient struct {
	mqClient mq_http_sdk.MQClient

	accessKey string // AccessKey 阿里云身份验证，在阿里云服务器管理控制台创建

	secretKey string // SecretKey

	endpoint string // 设置HTTP接入域名

	instanceId string //实例id

	logger log.Logger

	conf config.Config
}

func NewMQClient(options ...Option) *MQClient {
	poolOnce.Do(func() {
		onceClient = &MQClient{}
		for _, option := range options {
			option(onceClient)
		}

		if onceClient.conf == nil {
			onceClient.conf = config.NewNullConfig()
		}

		if onceClient.logger == nil {
			onceClient.logger = log.NewLogger()
		}

		onceClient.endpoint = onceClient.conf.GetString("aliyun_endpoint")
		if onceClient.endpoint == "" {
			onceClient.logger.Panicf("aliyun_endpoint is null! please confirm")
		}

		onceClient.accessKey = onceClient.conf.GetString("aliyun_access_key")
		if onceClient.accessKey == "" {
			onceClient.logger.Panicf("aliyun_access_key is null! please confirm")
		}

		onceClient.secretKey = onceClient.conf.GetString("aliyun_secrect_key")
		if onceClient.secretKey == "" {
			onceClient.logger.Panicf("aliyun_secrect_key is null! please confirm")
		}

		onceClient.instanceId = onceClient.conf.GetString("aliyun_instance_id")
		if onceClient.instanceId == "" {
			onceClient.logger.Panicf("aliyun_instance_id is null! please confirm")
		}

		onceClient.mqClient = mq_http_sdk.NewAliyunMQClient(onceClient.endpoint,
			onceClient.accessKey,
			onceClient.secretKey,
			"")

		onceClient.logger.Infof("[rocket mq] init success %s",
			onceClient.endpoint)
	})

	return onceClient
}

func (mc *MQClient) Consumer(topicName string, groupID string, messageTag string) mq_http_sdk.MQConsumer {
	return mc.mqClient.GetConsumer(
		mc.instanceId, topicName, groupID, messageTag)
}

func (mc *MQClient) Producer(topicName string) mq_http_sdk.MQProducer {
	return mc.mqClient.GetProducer(
		mc.instanceId, topicName)
}

type MQClientOptions struct{}

type Option func(*MQClient)

func (MQClientOptions) WithConf(conf config.Config) Option {
	return func(mq *MQClient) {
		mq.conf = conf
	}
}

func (MQClientOptions) WithLogger(logger log.Logger) Option {
	return func(mq *MQClient) {
		mq.logger = logger
	}
}
