package resp_code

// 支付状态
const (
	TRADE_SUCCESS  = "TRADE_SUCCESS"
	TRADE_FAIL     = "TRADE_FAIL"
	TRADE_TIMEOUT  = "TRADE_TIMEOUT"
	TRADE_WAIT_PAY = "TRADE_WAIT_PAY"
	TRADE_PROCESS  = "TRADE_PROCESS"
	TRADE_NOTFOUND = "TRADE_NOTFOUND"
)

// 订单状态
const (
	ORDER_INIT    = "0000000000"
	ORDER_SUCCESS = "1000000000"
	ORDER_FAIL    = "2000000000"
	ORDER_REJECT  = "3000000000"
	ORDER_TIMEOUT = "5000000000"
	ORDER_PROCESS = "6000000000"
)

func GetOrderST(tradeStatus string) string {
	var st string
	switch tradeStatus {
	case TRADE_SUCCESS: /*成功*/
		st = ORDER_SUCCESS
	case TRADE_PROCESS: /*处理中*/
		st = ORDER_PROCESS
	case TRADE_TIMEOUT: /*超时*/
		st = ORDER_TIMEOUT
	case TRADE_FAIL: /*失败*/
		st = ORDER_FAIL
	default: /*默认失败*/
		st = ORDER_FAIL
	}
	return st
}
