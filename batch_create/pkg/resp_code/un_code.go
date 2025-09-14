package resp_code

const (
	QUANQUDAO_SUCCESS          = "SUCCESS"          // 交易成功
	QUANQUDAO_PRESUMED_SUCCESS = "PRESUMED_SUCCESS" // 支付推定成功

	QUANQUDAO_PROCESSING = "PROCESSING" // 交易处理中
	QUANQUDAO_UNKNOWN    = "UNKNOWN"    // 交易结果未知

	QUANQUDAO_FAILURE = "FAILURE" // 交易失败
	QUANQUDAO_REFUSED = "REFUSED" // 订单拒绝支付

	QUANQUDAO_REFUNDED = "REFUNDED" // 已退货
	QUANQUDAO_REVERSED = "REVERSED" // 已冲正

	QUANQUDAO_PRE_AUTH_COMPLETED         = "PRE_AUTH_COMPLETED"         // 已被预授权完成
	QUANQUDAO_PRE_AUTH_CANCELED          = "PRE_AUTH_CANCELED"          // 已被预授权撤销
	QUANQUDAO_PRE_AUTH_COMPLETE_CANCELED = "PRE_AUTH_COMPLETE_CANCELED" // 已被预授
)

const (
	UN_CODE_SUCCESS = "0000000000" // 成功
)

var SuccessList []string

func init() {
	SuccessList = make([]string, 0)
	SuccessList = append(SuccessList, TRADE_SUCCESS, QUANQUDAO_SUCCESS, QUANQUDAO_PRESUMED_SUCCESS)
}
