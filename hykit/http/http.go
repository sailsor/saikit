package http

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	"code.jshyjdtech.com/godev/hykit/log"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	logger log.Logger

	transports http.Transport

	Client *resty.Client // go-resty用法灵活，大写公开，不必局限于该文件下的SendPOST、SendGET等方法；
}

type Options func(*Client)

type ClientOptions struct{}

func NewClient(opts ...Options) *Client {
	c := &Client{Client: resty.New(),
		transports: http.Transport{}}
	for _, opt := range opts {
		opt(c)
	}

	if c.logger == nil {
		c.logger = log.NewLogger()
	}

	c.Client.SetTransport(&c.transports)
	return c
}

// WithInsecureSkip with TLS/SSL
func (ClientOptions) WithInsecureSkip() Options {
	return func(hc *Client) {
		hc.transports.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
}

// WithDisableKeepAlive 禁用长连接
func (ClientOptions) WithDisableKeepAlive(b bool) Options {
	return func(hc *Client) {
		hc.transports.DisableKeepAlives = b
	}
}

// WithMaxIdleConn 空闲连接
func (ClientOptions) WithMaxIdleConn(n int) Options {
	return func(hc *Client) {
		hc.transports.MaxIdleConns = n
	}
}

func (ClientOptions) WithTimeOut(timeout time.Duration) Options {
	return func(hc *Client) {
		hc.Client.SetTimeout(timeout)
	}
}

// WithHttpProxyURL proxy
func (ClientOptions) WithHttpProxyURL(proxyURL string) Options {
	return func(hc *Client) {
		pURL, err := url.Parse(proxyURL)
		if err != nil {
			return
		}
		hc.transports.Proxy = http.ProxyURL(pURL)
	}
}

func (ClientOptions) WithLogger(logger log.Logger) Options {
	return func(hc *Client) {
		hc.logger = logger
	}
}
