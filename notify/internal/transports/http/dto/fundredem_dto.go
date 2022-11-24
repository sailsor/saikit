package dto

import (
	"notify/internal/domain"
	"notify/internal/infra"
)

type UnionNotify struct {
	infra              *infra.Infra
	ProdCd             string /*产品代码, 不参与签名*/
	Version            string `form:"version"`            // 版本号
	Encoding           string `form:"encoding"`           // 编码方式
	CertId             string `form:"certId"`             // 证书 ID
	Signature          string `form:"signature"`          // 签名
	SignMethod         string `form:"signMethod"`         // 签名方法
	TxnType            string `form:"txnType"`            // 交易类型
	TxnSubType         string `form:"txnSubType"`         // 交易子类
	BizType            string `form:"bizType"`            // 产品类型
	ChannelType        string `form:"channelType"`        // 渠道类型
	AccessType         string `form:"accessType"`         // 接入类型
	AcqInsCode         string `form:"acqInsCode"`         // 收单机构代码
	MerId              string `form:"merId"`              // 商户代码
	OrderId            string `form:"orderId"`            // 商户订单号
	TxnTime            string `form:"txnTime"`            // 订单发送时间
	AccNo              string `form:"accNo"`              // 账号
	TxnAmt             string `form:"txnAmt"`             // 交易金额
	CurrencyCode       string `form:"currencyCode"`       // 交易币种
	ReqReserved        string `form:"reqReserved"`        // 请求方保留域
	Reserved           string `form:"reserved"`           // 保留域
	QueryId            string `form:"queryId"`            // 查询流水号
	RespCode           string `form:"respCode"`           // 响应码
	RespMsg            string `form:"respMsg"`            // 应答信息
	UnionOrderId       string `form:"unionOrderId"`       // 原银联贷记订单号
	SettleDate         string `form:"settleDate"`         // 清算日期
	SettleCurrencyCode string `form:"settleCurrencyCode"` // 清算币种
	SettleAmt          string `form:"settleAmt"`          // 清算金额
	TraceNo            string `form:"traceNo"`            // 系统跟踪号
	TraceTime          string `form:"traceTime"`          // 交易传输时间
	ExchangeRate       string `form:"exchangeRate"`       // 清算汇率
	SignPubKeyCert     string `form:"signPubKeyCert"`     // 签名公钥证书
}

/*获取加密请求报文*/
func (un *UnionNotify) GetSignBuf() string {
	return domain.SignatureBuffer(un)
}
