package security

import (
	"crypto/rsa"
	"crypto/x509"
	"io/ioutil"

	"github.com/pkg/errors"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

type Cert struct {
	keyType string
	keyFile string
	keyPass string

	certificate *x509.Certificate
	rsaPrivate  *RSAKey
	rsaPublic   *RSAKey
}

func NewCert(cfg map[string]string) (*Cert, error) {
	var err error
	cert := new(Cert)
	cert.keyFile = cfg["KEY_FILE"]
	cert.keyType = cfg["KEY_TYPE"]
	cert.keyPass = cfg["KEY_PASS"] // PFX 密码

	keyBuf, err := ioutil.ReadFile(cert.keyFile)
	if err != nil {
		return nil, errors.WithMessagef(err, "读取证书文件失败[%s]", cert.keyFile)
	}

	switch cert.keyType {
	case FILE_CERT_CER:
		pemBlock, err := DecodePEM(keyBuf)
		if err != nil {
			return nil, errors.Errorf("PEM Decode failed[%s][%s]", keyBuf, err)
		}
		cert.certificate, err = x509.ParseCertificate(pemBlock.Bytes)
		if err != nil {
			return nil, errors.Errorf("解析PEM证书[%s]失败[%s]", cert.keyFile, err)
		}

		if pub, ok := cert.certificate.PublicKey.(*rsa.PublicKey); ok {
			cert.rsaPublic = loadFromRSAPublic(pub)
		} else {
			return nil, errors.Errorf("CERT公钥提取失败[%s]失败", cert.keyFile)
		}
	case FILE_CERT_PFX:
		var pri interface{}
		pri, cert.certificate, _, err = pkcs12.DecodeChain(keyBuf, cert.keyPass)
		if err != nil {
			return nil, errors.Errorf("PFX证书加载失败[%s]失败[%s]", cert.keyFile, err)
		}
		if p, ok := pri.(*rsa.PrivateKey); ok {
			cert.rsaPrivate = loadFromRSAPrivate(p)
			cert.rsaPublic = loadFromRSAPublic(cert.rsaPrivate.PublicKey)
		} else {
			return nil, errors.Errorf("FPX提取私钥失败[%s]失败", cert.keyFile)
		}
	default:
		return nil, errors.Errorf("非法类型[%s]失败", cert.keyType)
	}
	return cert, nil
}

func LoadCertPublicPEM(buf []byte) (*Cert, error) {
	var err error
	cert := new(Cert)

	pemBlock, err := DecodePEM(buf)
	if err != nil {
		return nil, errors.Errorf("PEM Decode failed[%s]err[%s]", buf, err)
	}

	cert.certificate, err = x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, errors.Errorf("解析PEM证书[%s]失败[%s]", cert.keyFile, err)
	}
	if pub, ok := cert.certificate.PublicKey.(*rsa.PublicKey); ok {
		cert.rsaPublic = loadFromRSAPublic(pub)
	} else {
		return nil, errors.Errorf("CERT 公钥提取失败[%s]失败", cert.keyFile)
	}
	cert.keyType = FILE_CERT_CER
	return cert, nil
}

func LoadCertPrivatePFX(buf []byte, pass string) (*Cert, error) {
	var err error
	cert := new(Cert)

	var pri interface{}
	pri, cert.certificate, _, err = pkcs12.DecodeChain(buf, pass)
	if err != nil {
		return nil, errors.Errorf("PFX证书加载失败[%s];", err)
	}
	if p, ok := pri.(*rsa.PrivateKey); ok {
		cert.rsaPrivate = loadFromRSAPrivate(p)
	} else {
		return nil, errors.Errorf("FPX私钥提取失败[%s];", err)
	}
	cert.keyType = FILE_CERT_PFX
	return cert, nil
}

func (cert *Cert) GetPublicRSAKey() (*RSAKey, error) {
	return cert.rsaPublic, nil
}

func (cert *Cert) GetPrivateRSAKey() (*RSAKey, error) {
	if cert.keyType == FILE_CERT_CER {
		return nil, errors.Errorf("公钥证书不支持提取私钥")
	}
	return cert.rsaPrivate, nil
}

/*Sign
   algo 算法:
  signBlock 签名串
  signature 签名值
*/
func (cert *Cert) Sign(algo int, signBlock []byte) ([]byte, error) {
	return cert.rsaPrivate.Sign(algo, signBlock)
}

// Verify
func (cert *Cert) Verify(algo int, signBlock, signature []byte) error {
	return cert.rsaPublic.Verify(algo, signBlock, signature)
}

// EncryptPKCS1v15 /* EncryptPKCS1v15: 公钥加密:
func (cert *Cert) EncryptPKCS1v15(plainBlock []byte) ([]byte, error) {
	return cert.rsaPublic.RSAPublicEncryptPKCS1v15(plainBlock)
}

// DecryptPKCS1v15 /*DecryptPKCS1v15: 私钥解密*/
func (cert *Cert) DecryptPKCS1v15(cipherBlock []byte) ([]byte, error) {
	return cert.rsaPrivate.RSAPrivateDecryptPKCS1v15(cipherBlock)
}

/*SerialNumber 获取证书序列号.*/
func (cert *Cert) SerialNumber() string {
	return cert.certificate.SerialNumber.String()
}

/*SerialNumberHex 获取证书序列号十六进制（用于windows查看证书序列号为十六进制情况）.*/
func (cert *Cert) SerialNumberHex() string {
	return string(EncodeHex(cert.certificate.SerialNumber.Bytes()))
}

/*Certificate 获取证书信息.*/
func (cert *Cert) Certificate() *x509.Certificate {
	return cert.certificate
}
