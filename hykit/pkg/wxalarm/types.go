package wxalarm

const (
	WxHookKey          = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send"
	WxErrCodeFreqLimit = 45009
)

var (
	AlarmTemplate = `异常消息:
- 环境: %s
- 服务名称: %s
- 消息IP: %s
- 消息时间: %s
- 提示: %s
- 错误信息: %v`

	NotifyTemplate = `通知消息:
- 子系统: %s
- 消息时间: %s
- 主题: %s
- 详情: %v`
)

type Config struct {
	WebHook  string `json:"web_hook"`
	Retries  int    `json:"retries"`
	Interval int    `json:"interval"` // second
}

type Output struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

func (w *WXAlarm) setDefaultConfig() {
	w.config.WebHook = w.conf.GetString("wx_web_hook")
	if w.config.WebHook == "" {
		w.logger.Panicf("wx web_hook is empty!")
	}

	w.config.Interval = w.conf.GetInt("wx_interval")
	if w.config.Interval == 0 {
		w.config.Interval = 1
	}

	w.config.Retries = w.conf.GetInt("wx_retries")
	if w.config.Retries == 0 {
		w.config.Retries = 3
	}
}
