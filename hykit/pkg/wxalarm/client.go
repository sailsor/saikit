package wxalarm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"code.jshyjdtech.com/godev/hykit/log"

	"code.jshyjdtech.com/godev/hykit/config"
)

type WXAlarm struct {
	conf   config.Config
	logger log.Logger
	config Config
}

type Option func(*WXAlarm)

func New(opts ...Option) *WXAlarm {
	wx := &WXAlarm{}
	for _, opt := range opts {
		opt(wx)
	}

	if wx.conf == nil {
		wx.conf = config.NewMemConfig()
	}

	wx.setDefaultConfig()

	return wx
}

func WithConf(conf config.Config) Option {
	return func(wx *WXAlarm) {
		wx.conf = conf
	}
}

func WithLogger(logger log.Logger) Option {
	return func(wx *WXAlarm) {
		wx.logger = logger
	}
}

func (w *WXAlarm) SendMessage(ctx context.Context, message string) (*Output, error) {
	url := fmt.Sprintf("%s?key=%s", WxHookKey, w.config.WebHook)
	resp, err := http.Post(url, "application/json", strings.NewReader(message))
	if err != nil {
		w.logger.Errorc(ctx, "Send wx message error:%v, content:%v", err, message)
		return nil, err
	}
	defer func() {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	output := &Output{}
	err = json.NewDecoder(resp.Body).Decode(output)
	if err != nil {
		w.logger.Errorc(ctx, "Send wx message error:%v, output:%v", err, output)
		return nil, err
	}

	return output, nil
}

// 发送text形式消息
func (w *WXAlarm) SendTextMessage(ctx context.Context, message string, mentionedList, MentionedMobileList []string) (*Output, error) {
	var body struct {
		MsgType string `json:"msgtype"`
		Text    struct {
			Content             string   `json:"content"`
			MentionedList       []string `json:"mentioned_list,omitempty"`
			MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
		} `json:"text"`
	}
	body.MsgType = "text"
	body.Text.Content = message
	body.Text.MentionedList = mentionedList
	body.Text.MentionedMobileList = MentionedMobileList

	data, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}

	return w.SendMessage(ctx, string(data))
}

// 发送markdown形式消息
func (w *WXAlarm) SendMarkdownMessage(ctx context.Context, message string) (*Output, error) {
	var body struct {
		MsgType  string `json:"msgtype"`
		Markdown struct {
			Content string `json:"content"`
		} `json:"markdown"`
	}
	body.MsgType = "markdown"
	body.Markdown.Content = message

	data, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}

	return w.SendMessage(ctx, string(data))
}

// 发送失败进行重试，只限于网络报错或者频率限制错误进行重试
// 谨慎使用，该函数会阻塞调用
func (w *WXAlarm) SendTextMessageWithRetry(ctx context.Context, message string, mentionedList, MentionedMobileList []string) (output *Output, err error) {
	for i := 0; i < w.config.Retries; i++ {
		output, err = w.SendTextMessage(ctx, message, mentionedList, MentionedMobileList)
		if err != nil || output.ErrCode == WxErrCodeFreqLimit { // 网络报错或者频率限制错误进行重试
			time.Sleep(time.Duration(w.config.Interval) * time.Second)
			continue
		}
		break
	}
	return
}

// 发送失败进行重试，只限于网络报错或者频率限制错误进行重试
// 谨慎使用，该函数会阻塞调用
func (w *WXAlarm) SendMarkdownMessageWithRetry(ctx context.Context, message string) (output *Output, err error) {
	for i := 0; i < w.config.Retries; i++ {
		output, err = w.SendMarkdownMessage(ctx, message)
		if err != nil || output.ErrCode == WxErrCodeFreqLimit { // 网络报错或者频率限制错误进行重试
			time.Sleep(time.Duration(w.config.Interval) * time.Second)
			continue
		}
		break
	}
	return
}
