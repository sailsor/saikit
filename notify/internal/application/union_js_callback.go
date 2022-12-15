package application

import (
	"code.jshyjdtech.com/godev/hykit/pkg/security"
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"notify/internal/domain/entity"
	"notify/internal/infra"
	"notify/internal/transports/http/dto"
	"time"
)

/*
江苏银联异步回调
*/
type UnionJsCallbackSvc struct {
	*infra.Infra
	tranRespTopic string
	UnNotify      *dto.UnionNotify
}

func NewUnionJsCallbackSvc(inf *infra.Infra) *UnionJsCallbackSvc {
	cb := new(UnionJsCallbackSvc)
	cb.Infra = inf
	cb.UnNotify = new(dto.UnionNotify)
	return cb
}

func (cb *UnionJsCallbackSvc) VerifySign(ctx context.Context) error {
	logger := cb.Logger
	logger.Infoc(ctx, "VerifySign:开始调用；")
	// 组签名block
	signature := security.DecodeBase64([]byte(cb.UnNotify.Signature))
	signBuf := cb.UnNotify.GetSignBuf()

	keyBuf := cb.UnNotify.SignPubKeyCert
	if keyBuf == "" {
		logger.Errorc(ctx, "提取响应公钥失败[%s]", keyBuf)
		return errors.Errorf("提取响应公钥失败[%s]", keyBuf)
	}

	switch cb.UnNotify.SignMethod {
	case "01":
		pubCert, err := security.LoadCertPublicPEM([]byte(keyBuf))
		if err != nil {
			logger.Errorc(ctx, "解析公钥失败[%s]", keyBuf)
			return errors.Errorf("解析公钥失败[%s]", err)
		}
		hexBuf, err := BuildHashHex(ctx, security.SHA256WithRSA, []byte(signBuf))
		if err != nil {
			logger.Errorc(ctx, "BuildHashHex失败[%s]", err)
			return err
		}
		err = pubCert.Verify(security.SHA256WithRSA, hexBuf, signature)
		if err != nil {
			logger.Errorc(ctx, "RSA验签失败[%s]", err)
			return err
		}
		logger.Infoc(ctx, "RSA验签成功")

	case "02":
		pubCert, err := security.LoadECDSACertPublicCER([]byte(keyBuf))
		if err != nil {
			logger.Errorc(ctx, "解析公钥失败[%s]", keyBuf)
			return errors.Errorf("解析公钥失败[%s]", err)
		}

		hexBuf, err := BuildSm3Hex(ctx, []byte(signBuf))
		if err != nil {
			logger.Errorc(ctx, "BuildHashHex失败[%s]", err)
			return err
		}

		err = pubCert.Verify(hexBuf, signature)
		if err != nil {
			logger.Errorc(ctx, "SM2验签失败[%s]", err)
			return err
		}
		logger.Infoc(ctx, "SM2验签成功")

	default:
		logger.Errorc(ctx, "非法的签名类型[%s]", cb.UnNotify.SignMethod)
		return errors.Errorf("非法的签名类型[%s]", cb.UnNotify.SignMethod)
	}
	logger.Infoc(ctx, "VerifySign:结束调用；")
	return nil
}

func BuildHashHex(ctx context.Context, algo int, data []byte) ([]byte, error) {
	hashData, err := security.Hash(algo, data)
	if err != nil {
		return nil, err
	}
	result := security.EncodeHex(hashData)
	return result, nil
}

func BuildSm3Hex(ctx context.Context, data []byte) ([]byte, error) {

	return security.EncodeHex(security.Sm3hash(data)), nil
}

func (cb *UnionJsCallbackSvc) Record(ctx context.Context) error {
	logger := cb.Logger
	logger.Infoc(ctx, "Record:开始调用；")

	tblNotifyItem := new(entity.TblNotifyItem)
	tblNotifyItem.AppId = cb.UnNotify.AcqInsCode
	tblNotifyItem.OrderId = cb.UnNotify.OrderId
	tblNotifyItem.TradeId = cb.UnNotify.UnionOrderId
	tblNotifyItem.TranTime = time.Now().Format("2006-01-02 15:04:05")

	tx := cb.Infra.DB.GetDb("appdb").Begin(&sql.TxOptions{})
	defer tx.Rollback()
	err := tx.Create(tblNotifyItem).Error
	if err != nil {
		logger.Errorc(ctx, "数据库插入失败[%s]", err)
	}
	tx.Commit()

	logger.Infoc(ctx, "Record:结束调用；")
	return nil
}
