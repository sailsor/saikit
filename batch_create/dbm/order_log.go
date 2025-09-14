package dbm

import (
	"batch_create/pkg/resp_code"
	"batch_create/pkg/sequence"
	"batch_create/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type TblProdOrderLog struct {
	ID int64 `json:"id,omitempty" gorm:"column:id;primary_key"`

	// 本地日期
	LocalDate string `json:"localDate,omitempty" gorm:"column:local_date"`

	// 本地时间
	LocalTime string `json:"localTime,omitempty" gorm:"column:local_time"`

	// 机构号
	AppID string `json:"appId,omitempty" gorm:"column:app_id"`

	// 产品码
	ProdCd string `json:"prodCd,omitempty" gorm:"column:prod_cd"`

	// 交易码
	TranCd string `json:"tranCd,omitempty" gorm:"column:tran_cd"`

	// 清算日期
	HostDate string `json:"hostDate,omitempty" gorm:"column:host_date"`

	// 订单号
	OrderID string `json:"orderId,omitempty" gorm:"column:order_id"`

	// 批次号
	BatchNo string `json:"batchNo,omitempty" gorm:"column:batch_no"`

	// 订单状态
	OrderSt string `json:"orderSt,omitempty" gorm:"column:order_st"`

	// 交易号
	TradeID string `json:"tradeId,omitempty" gorm:"column:trade_id"`

	// 交易状态
	TradeStatus string `json:"tradeStatus,omitempty" gorm:"column:trade_status"`

	// 入账状态
	EntryStatus string `json:"entryStatus,omitempty" gorm:"column:entry_status"`

	// 账务状态
	AcctStatus string `json:"acctStatus,omitempty" gorm:"column:acct_status"`

	// 应答码
	RespCode string `json:"respCode,omitempty" gorm:"column:resp_code"`

	// 应答信息
	RespMsg string `json:"respMsg,omitempty" gorm:"column:resp_msg"`

	// 详细应答码
	SubRespCode string `json:"subRespCode,omitempty" gorm:"column:sub_resp_code"`

	// 详细应答信息
	SubRespMsg string `json:"subRespMsg,omitempty" gorm:"column:sub_resp_msg"`

	// 交易时间
	TranTime string `json:"tranTime,omitempty" gorm:"column:tran_time"`

	// 交易金额
	TranAmt string `json:"tranAmt,omitempty" gorm:"column:tran_amt"`

	// 会员手续费
	TranFee string `json:"tranFee,omitempty" gorm:"column:tran_fee"`

	// 商户号
	MchtCd string `json:"mchtCd,omitempty" gorm:"column:mcht_cd"`

	// 商户名称
	MchtName string `json:"mchtName,omitempty" gorm:"column:mcht_name"`

	// 会员号
	UserID string `json:"userId,omitempty" gorm:"column:user_id"`

	// 会员姓名
	UserName string `json:"userName,omitempty" gorm:"column:user_name"`

	// 手机号
	PhoneNo string `json:"phoneNo,omitempty" gorm:"column:phone_no"`

	// 证件号码
	CertifID string `json:"certifId,omitempty" gorm:"column:certif_id"`

	// 支付卡类型
	PriAcctTp string `json:"priAcctTp,omitempty" gorm:"column:pri_acct_tp"`

	// 付款方开户行行号
	//PriBankNo string `json:"priBankNo,omitempty" gorm:"column:pri_bank_no"`

	// 付款方银行账号
	PriBankAcct string `json:"priBankAcct,omitempty" gorm:"column:pri_bank_acct"`

	// 支付卡号
	PriAcctNo string `json:"priAcctNo,omitempty" gorm:"column:pri_acct_no"`

	// 付款方账户名称
	PriAcctNm string `json:"priAcctNm,omitempty" gorm:"column:pri_acct_nm"`

	// 收款方介质
	PyeAcctOp string `json:"pyeAcctOp,omitempty" gorm:"column:pye_acct_op"`

	// 收款方类型
	PyeAcctTp string `json:"pyeAcctTp,omitempty" gorm:"column:pye_acct_tp"`

	// 收款方账号
	PyeAcctNo string `json:"pyeAcctNo,omitempty" gorm:"column:pye_acct_no"`

	// 收款方姓名
	PyeAcctNm string `json:"pyeAcctNm,omitempty" gorm:"column:pye_acct_nm"`

	// 支付token
	PyeToken string `json:"pyeToken,omitempty" gorm:"column:pye_token"`

	// 收款方开户行行号
	PyeBankNo string `json:"pyeBankNo,omitempty" gorm:"column:pye_bank_no"`

	// 付款银行账号密文
	CipPriBankAcct string `json:"cipPriBankAcct,omitempty" gorm:"column:cip_pri_bank_acct"`

	// 支付卡号密文
	CipPriAcctNo string `json:"cipPriAcctNo,omitempty" gorm:"column:cip_pri_acct_no"`

	// 支付会员姓名
	CipPriAcctNm string `json:"cipPriAcctNm,omitempty" gorm:"column:cip_pri_acct_nm"`

	// 收款方账号-密文
	CipPyeAcctNo string `json:"CipPyeAcctNo,omitempty" gorm:"column:cip_pye_acct_no"`

	// 收款方姓名-密文
	CipPyeAcctNm string `json:"CipPyeAcctNm,omitempty" gorm:"column:cip_pye_acct_nm"`

	// 协议号
	ProtocolID string `json:"protocolId,omitempty" gorm:"column:protocol_id"`

	// 附言
	Purpose string `json:"purpose,omitempty" gorm:"column:purpose"`

	// 保留信息
	Reserve string `json:"reserve,omitempty" gorm:"column:reserve"`

	// 卡类型
	CardType string `json:"cardType,omitempty" gorm:"column:card_type"`

	// 卡bin
	CardBin string `json:"cardBin,omitempty" gorm:"column:card_bin"`

	// 发卡机构号
	IssInsID string `json:"issInsId,omitempty" gorm:"column:iss_ins_id"`

	// 发卡机构名称
	IssInsName string `json:"issInsName,omitempty" gorm:"column:iss_ins_name"`

	// 资金来源
	SourceFund string `json:"sourceFund,omitempty" gorm:"column:source_fund"`

	// 产品类型
	ProdType string `json:"prodType,omitempty" gorm:"column:prod_type"`

	// 渠道交易号
	ChnlTradeID string `json:"chnlTradeId,omitempty" gorm:"column:chnl_trade_id"`

	// 银联交易号
	UnionOrderId string `json:"unionOrderId,omitempty" gorm:"column:union_order_id"`

	// 渠道交易时间
	ChnlTranTime string `json:"chnlTranTime,omitempty" gorm:"column:chnl_tran_time"`

	// 渠道机构号
	ChnlServerID string `json:"chnlServerId,omitempty" gorm:"column:chnl_server_id"`

	// 渠道商户号
	ChnlMchntCd string `json:"chnlMchntCd,omitempty" gorm:"column:chnl_mchnt_cd"`

	// 渠道商户名称
	ChnlMchntName string `json:"chnlMchntName,omitempty" gorm:"column:chnl_mchnt_name"`

	// 渠道终端号
	ChnlTermID string `json:"chnlTermId,omitempty" gorm:"column:chnl_term_id"`

	// 结算模式
	SettMd string `json:"settMd,omitempty" gorm:"column:sett_md"`

	InsFee    string `json:"insFee,omitempty" gorm:"column:ins_fee"`
	ExpFee    string `json:"expFee,omitempty" gorm:"column:exp_fee"`
	SettFee   string `json:"settFee,omitempty" gorm:"column:sett_fee"`
	LocalFee  string `json:"localFee,omitempty" gorm:"column:local_fee"`
	ChanelFee string `json:"chanelFee,omitempty" gorm:"column:chanel_fee"`

	// 异步回调地址
	NotifyURL string `json:"notifyUrl,omitempty" gorm:"column:notify_url"`

	// 风控设备号
	RiskDevID string `json:"riskDevId,omitempty" gorm:"column:risk_dev_id"`

	// 风控设备类型
	RiskDevTp string `json:"riskDevTp,omitempty" gorm:"column:risk_dev_tp"`

	// 风控手机号
	RiskPhone string `json:"riskPhone,omitempty" gorm:"column:risk_phone"`

	// 风控IP
	RiskIP string `json:"riskIp,omitempty" gorm:"column:risk_ip"`

	// 风控经纬度
	RiskLocation string `json:"riskLocation,omitempty" gorm:"column:risk_location"`

	// 风控Sim卡
	RiskSimNo string `json:"riskSimNo,omitempty" gorm:"column:risk_sim_no"`

	// 风控设备SIM卡数量
	RiskSimCnt string `json:"riskSimCnt,omitempty" gorm:"column:risk_sim_cnt"`

	// 风控账号id
	RiskAcctId string `json:"riskAcctId,omitempty" gorm:"column:risk_acct_id"`

	// 风控设备名
	RiskDevName string `json:"riskDevName,omitempty" gorm:"column:risk_dev_name"`

	// 风控设备评级
	RiskDevStore string `json:"riskDevStore,omitempty" gorm:"column:risk_dev_store"`

	// 对账状态
	SettSt string `json:"settSt,omitempty" gorm:"column:sett_st"`

	// 渠道键值
	OutSettKey string `json:"outSettKey,omitempty" gorm:"column:out_sett_key"`

	// 渠道清算日期
	OutHostDate string `json:"outHostDate,omitempty" gorm:"column:out_host_date"`

	CreatedAt time.Time `json:"createdAt,omitempty" gorm:"column:created_at;default:'CURRENT_TIMESTAMP(6)'"`

	UpdatedAt time.Time `json:"updatedAt,omitempty" gorm:"column:updated_at;default:'CURRENT_TIMESTAMP(6)'"`
}

func (od *TblProdOrderLog) TableName() string {
	return "tbl_prod_order_log"
}

func (od *TblProdOrderLog) GetNoLock(tx *gorm.DB) error {
	logger := utils.GlobalLogger

	r := tx.Where(od).First(od)
	if r.Error == gorm.ErrRecordNotFound {
		logger.Errorf("没有找到原交易信息:%s", r.Error.Error())
		return r.Error
	} else if r.Error != nil {
		logger.Errorf("查找原交易失败[%s];", r.Error)
		return r.Error
	}
	return nil
}

func (od *TblProdOrderLog) GetLk(tx *gorm.DB) error {
	logger := utils.GlobalLogger

	r := tx.Set("gorm:query_option", "FOR UPDATE").
		Where(TblProdOrderLog{OrderID: od.OrderID, TradeID: od.TradeID}).
		First(od)
	if r.Error == gorm.ErrRecordNotFound {
		logger.Errorf("没有找到原交易信息:%s", r.Error.Error())
		return r.Error
	} else if r.Error != nil {
		logger.Errorf("查找原交易失败[%s];", r.Error)
		return r.Error
	}
	return nil
}

func (od *TblProdOrderLog) Update(tx *gorm.DB) error {
	logger := utils.GlobalLogger

	/*更新订单明细*/
	if od.ID == 0 {
		return errors.Errorf("主键ID为空")
	}
	r := tx.Model(od).Updates(od)
	if r.Error != nil {
		logger.Errorf("更新订单:订单号[%s]交易号[%s]失败[%s];",
			od.OrderID, od.TradeID, r.Error)
		return errors.Errorf("更新订单失败")
	}
	return nil
}

func (od *TblProdOrderLog) Create(tx *gorm.DB) error {
	logger := utils.GlobalLogger
	/*新增订单*/
	r := tx.Create(od)
	if r.Error != nil {
		logger.Errorf("订单插入失败[%v][%s][%s]", r.Error.Error())
		return r.Error
	}
	return nil
}

func NewOrder() *TblProdOrderLog {
	now := time.Now()
	order := new(TblProdOrderLog)
	order.LocalDate = now.Format("20060102")
	order.LocalTime = now.Format("150405")
	order.OrderSt = resp_code.ORDER_INIT
	order.TradeStatus = resp_code.TRADE_WAIT_PAY
	order.AppID = "TEST_APP"
	tradeId := sequence.GenTradeID()
	order.TradeID = tradeId
	order.OrderID = tradeId
	order.RespCode = resp_code.SUCCESS
	order.RespMsg = "受理中"
	order.SubRespCode = "00"
	order.SubRespMsg = "受理中"
	order.TranAmt = "100"
	order.MchtCd = "123081111111"
	order.MchtName = "测试商户"
	order.ProdCd = "FUNDREDEM"
	order.TranCd = "1200"
	order.HostDate = order.LocalDate
	return order
}
