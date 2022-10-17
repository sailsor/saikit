package supermng

import (
	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	"strings"
	"sync"
)

var clientOnce sync.Once

var onceClient *Client

type Client struct {
	//superClients map[string]*SuperClient
	superClients sync.Map
	superConfigs []SuperConfig
	conf         config.Config
	logger       log.Logger
}

type SuperConfig struct {
	SuperName string `json:"superName" yaml:"superName"`

	Host string `json:"host" yaml:"host"`

	Port string `json:"port" yaml:"port"`

	User string `json:"user" yaml:"user"`

	Password string `json:"password" yaml:"password"`

	LogPath string `json:"logPath" yaml:"logPath"`
}

type Option func(c *Client)

type ClientOptions struct{}

func (ClientOptions) WithConf(conf config.Config) Option {
	return func(r *Client) {
		r.conf = conf
	}
}

func (ClientOptions) WithLogger(logger log.Logger) Option {
	return func(r *Client) {
		r.logger = logger
	}
}

func NewClient(options ...Option) *Client {
	clientOnce.Do(func() {
		onceClient = &Client{
			superConfigs: make([]SuperConfig, 0),
		}

		for _, option := range options {
			option(onceClient)
		}

		if onceClient.conf == nil {
			onceClient.conf = config.NewNullConfig()
		}

		if onceClient.logger == nil {
			onceClient.logger = log.NewLogger()
		}

		onceClient.init()
	})

	return onceClient
}

func (c *Client) init() {
	superConfigs := make([]SuperConfig, 0)
	err := c.conf.UnmarshalKey("super_configs", &superConfigs)
	if err != nil {
		c.logger.Panicf("Fatal error config file: %s \n", err.Error())
	}

	c.superConfigs = superConfigs
	for _, superConfig := range superConfigs {
		sc, err := NewSuperClient(&superConfig)
		if err == nil {
			c.SetSuper(superConfig.SuperName, sc)
		}
	}
}

func (c *Client) SetSuper(superName string, sc *SuperClient) {
	superName = strings.ToLower(superName)
	c.superClients.Store(superName, sc)
}

func (c *Client) GetSuperClient(superName string) *SuperClient {

	superName = strings.ToLower(superName)
	if sc, ok := c.superClients.Load(superName); ok {
		v, o := sc.(*SuperClient)
		if !o {
			c.logger.Errorf("Type Assertion Fail ...")
			return nil
		}
		return v
	}
	c.logger.Errorf("[%s] not found", superName)
	return nil
}
