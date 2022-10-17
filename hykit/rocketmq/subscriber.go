package rocketmq

import (
	"context"
	"strings"
	"sync"
	"time"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	mq_http_sdk "github.com/aliyunmq/mq-http-go-sdk"
	"github.com/gogap/errors"
)

var subscriberOnce sync.Once
var onceSubEngine *SubscribeEngine

type ConsumeMessage struct {
	Messages mq_http_sdk.ConsumeMessageEntry
}

type ConsumeMessageAck struct {
	ReceiptHandle string
}

// HandlerFunc defines the handler used by rocketmq middleware as return value.
type HandlerFunc func(*Context)

// HandlersChain defines a HandlerFunc array.
type HandlersChain []HandlerFunc

// Last returns the last handler in the chain. ie. the last handler is the main one.
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

type Subscriber struct {
	topicName     string
	groupID       string
	messageTag    string
	messageNum    int
	maxWait       int
	concurrency   int
	handlersChain HandlersChain
}

type SubscribeEngine struct {
	client *MQClient

	logger log.Logger

	conf config.Config

	handlers HandlersChain

	list []*Subscriber

	pool sync.Pool
}

type SubscribeEngineOption func(*SubscribeEngine)

func NewSubscribeEngine(options ...SubscribeEngineOption) *SubscribeEngine {
	subscriberOnce.Do(func() {
		onceSubEngine = &SubscribeEngine{}

		for _, option := range options {
			option(onceSubEngine)
		}

		if onceSubEngine.conf == nil {
			onceSubEngine.conf = config.NewNullConfig()
		}

		if onceSubEngine.logger == nil {
			onceSubEngine.logger = log.NewLogger()
		}

		if onceSubEngine.handlers == nil {
			onceSubEngine.handlers = make(HandlersChain, 0)
		}
		if onceSubEngine.list == nil {
			onceSubEngine.list = make([]*Subscriber, 0)
		}

		/*初始化客户端*/
		var cliOpt MQClientOptions
		onceSubEngine.client = NewMQClient(
			cliOpt.WithLogger(onceSubEngine.logger),
			cliOpt.WithConf(onceSubEngine.conf))

		onceSubEngine.pool.New = func() interface{} {
			return onceSubEngine.allocateContext()
		}

	})

	return onceSubEngine
}

func WithSubscribeEngineConf(conf config.Config) SubscribeEngineOption {
	return func(se *SubscribeEngine) {
		se.conf = conf
	}
}

func WithSubscribeEngineLogger(logger log.Logger) SubscribeEngineOption {
	return func(se *SubscribeEngine) {
		se.logger = logger
	}
}

func (se *SubscribeEngine) allocateContext() *Context {
	return &Context{engine: se}
}

func (se *SubscribeEngine) Use(middleware ...HandlerFunc) {
	se.handlers = append(se.handlers, middleware...)
}

type SubscribeOption func(*Subscriber)

func (se *SubscribeEngine) Subscriber(options ...SubscribeOption) error {
	sub := new(Subscriber)
	for _, opt := range options {
		opt(sub)
	}

	if sub.topicName == "" {
		return errors.Errorf("topicName[%s]非法", sub.topicName)
	}
	if sub.groupID == "" {
		return errors.Errorf("groupID[%s]非法", sub.groupID)
	}
	//默认消息条数
	if sub.messageNum == 0 {
		sub.messageNum = 3
	}

	//默认轮询时间
	if sub.maxWait == 0 {
		sub.maxWait = 10
	}

	//默认并发
	if sub.concurrency == 0 {
		sub.concurrency = 1
	}
	//处理链
	sub.handlersChain = se.combineHandlers(sub.handlersChain)

	se.list = append(se.list, sub)

	se.logger.Infof("Subscriber[%+v] regitster success!", sub)
	return nil
}

func WithSubscribeTopicName(topicName string) SubscribeOption {
	return func(subscribe *Subscriber) {
		subscribe.topicName = topicName
	}
}

func WithSubscribeGroupID(groupID string) SubscribeOption {
	return func(subscribe *Subscriber) {
		subscribe.groupID = groupID
	}
}

func WithSubscribeMessageTag(messageTag string) SubscribeOption {
	return func(subscribe *Subscriber) {
		subscribe.messageTag = messageTag
	}
}

func WithSubscribeMessageNum(messageNum int) SubscribeOption {
	return func(subscribe *Subscriber) {
		subscribe.messageNum = messageNum
	}
}

func WithSubscribeMaxWait(maxWait int) SubscribeOption {
	return func(subscribe *Subscriber) {
		subscribe.maxWait = maxWait
	}
}

func WithSubscribeConcurrency(curr int) SubscribeOption {
	return func(subscribe *Subscriber) {
		subscribe.concurrency = curr
	}
}

func WithSubscribeHandlersChain(handlers ...HandlerFunc) SubscribeOption {
	return func(subscribe *Subscriber) {
		subscribe.handlersChain = handlers
	}
}

func (se *SubscribeEngine) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(se.handlers) + len(handlers)
	if finalSize >= int(abortIndex) {
		panic("too many handlers")
	}
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, se.handlers)
	copy(mergedHandlers[len(se.handlers):], handlers)
	return mergedHandlers
}

func (se *SubscribeEngine) handleConsumeMsg(subInfo *Subscriber, msg *ConsumeMessage, ack *ConsumeMessageAck) error {
	c := se.pool.Get().(*Context)
	defer se.pool.Put(c)

	c.reset()
	c.consumeMsg = msg
	c.consumeAck = ack
	c.subscriber = subInfo
	c.handlers = subInfo.handlersChain

	//调用链
	c.Next()

	if c.Errors != nil || c.IsAborted() {
		//处理失败
		return errors.Errorf("handleConsumeMsg 处理失败[%s][%v]", c.Errors, c.IsAborted())
	}

	if c.IsAck() {
		ack.ReceiptHandle = c.consumeAck.ReceiptHandle
		return nil
	}

	return nil
}

func (se *SubscribeEngine) Start() {

	var consumerFunc = func(consumer mq_http_sdk.MQConsumer, sub *Subscriber) {
		for {
			respChan := make(chan mq_http_sdk.ConsumeMessageResponse)
			errChan := make(chan error)
			//endChan := make(chan struct{})
			var wg sync.WaitGroup
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer cancel()

			wg.Add(1)
			go func(ctx context.Context) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					se.logger.Infof("ConsumeMessage 读取消息超时，返回重试；")
				case resp := <-respChan:
					// 处理业务逻辑
					var handles []string
					se.logger.Debugf("Consume [%d] messages---->", len(resp.Messages))

					for _, v := range resp.Messages {
						se.logger.Infof("Receive MessageID: [%s], PublishTime: [%d], MessageTag: [%s] "+
							" ConsumedTimes: [%d], FirstConsumeTime: [%d], NextConsumeTime: [%d]",
							v.MessageId, v.PublishTime, v.MessageTag, v.ConsumedTimes,
							v.FirstConsumeTime, v.NextConsumeTime)
						se.logger.Debugf("Receive MessageID: [%s], PublishTime: [%d], MessageTag: [%s], MessageBody[%s], MessageKey[%v]",
							v.MessageId, v.PublishTime, v.MessageTag, v.MessageBody, v.MessageKey)

						consumerMsg := &ConsumeMessage{v}
						consumerAck := &ConsumeMessageAck{}
						err := se.handleConsumeMsg(sub, consumerMsg, consumerAck)
						if err != nil {
							se.logger.Errorf("handleConsumeMsg 处理失败[%s][%s]下次接收时间[%d]",
								v.Message, v.MessageBody, v.NextConsumeTime)
							continue
						}

						handles = append(handles, consumerAck.ReceiptHandle)
					}

					// NextConsumeTime前若不确认消息消费成功，则消息会重复消费
					// 消息句柄有时间戳，同一条消息每次消费拿到的都不一样
					if len(handles) > 0 {
						ackErr := consumer.AckMessage(handles)
						if ackErr != nil {
							// 某些消息的句柄可能超时了会导致确认不成功
							se.logger.Errorf("AckMessage [%s]确认失败[%s]", handles, ackErr)
							for _, errAckItem := range ackErr.(errors.ErrCode).Context()["Detail"].([]mq_http_sdk.ErrAckItem) {
								se.logger.Errorf("AckMessage: ErrorHandle:%s, ErrorCode:%s, ErrorMsg:%s",
									errAckItem.ErrorHandle, errAckItem.ErrorCode, errAckItem.ErrorMsg)
							}
							se.logger.Infof("休眠3秒钟后继续....")
							time.Sleep(time.Duration(3) * time.Second)
						}
						se.logger.Infof("Ack ---->[%s]确认成功;", handles)
					}
				case err := <-errChan:
					if strings.Contains(err.(errors.ErrCode).Error(), "MessageNotExist") {
						se.logger.Debugf("ConsumeMessage No new message, continue")
						return
					}
					se.logger.Infof("ConsumeMessage 读取消息失败[%s]", err)
					se.logger.Infof("休眠2秒钟后继续....")
					time.Sleep(2 * time.Second)
				case <-time.After(40 * time.Second):
					se.logger.Infof("ConsumeMessage 读取消息超时，重试；")
				}
				return
			}(ctx)

			// 长轮询消费消息
			// 长轮询表示如果topic没有消息则请求会在服务端挂住3s，3s内如果有消息可以消费则立即返回
			// 一次最多消费3条(最多可设置为16条)
			// 长轮询时间3秒（最多可设置为30秒）
			go func() {
				consumer.ConsumeMessage(respChan, errChan, int32(sub.messageNum), int64(sub.maxWait))
				cancel()
			}()
			wg.Wait()
		}
	}

	if len(se.list) == 0 {
		se.logger.Panicf("没有订阅任何信息!!!")
	}

	//订阅配置对应队列
	for _, si := range se.list {
		se.logger.Infof("begin subscribe [%+v]", si)
		for i := 0; i < si.concurrency; i++ {
			go func(id int, s *Subscriber) {
				se.logger.Infof("[%d]订阅信息[%+v]", id+1, s)
				consumer := se.client.Consumer(s.topicName, s.groupID, s.messageTag)
				consumerFunc(consumer, s)
			}(i, si)
		}
	}
	se.logger.Infof("SubscribeEngine init success!")

}
