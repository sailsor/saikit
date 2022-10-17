package security

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"github.com/pkg/errors"
	p12 "github.com/tjfoc/gmsm/pkcs12"
	"github.com/tjfoc/gmsm/sm2"
	g509 "github.com/tjfoc/gmsm/x509"
)

type ECDSACert struct {
	KeyType     string `json:"key_type"`
	KeyFile     string `json:"key_file"`
	certificate *g509.Certificate
	publicKey   *sm2.PublicKey
	privateKey  *sm2.PrivateKey
}

/*
椭圆曲线SM2 pfx
*/
func LoadECDSACertPrivatePFX(buf []byte, pass string) (*ECDSACert, error) {
	var err error
	cert := new(ECDSACert)

	cert.certificate, cert.privateKey, err = decodePfxAllInfo(buf, pass)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

/*
椭圆曲线SM2 cer
*/
func LoadECDSACertPublicCER(buf []byte) (*ECDSACert, error) {
	var err error
	cert := new(ECDSACert)

	data, _ := pem.Decode(buf)
	if data == nil {
		return nil, errors.Errorf("pem decode fail, data is nil")
	}

	cert.certificate, err = g509.ParseCertificate(data.Bytes)
	if err != nil {
		return nil, err
	}

	// 使用 编组公钥x509.MarshalPKIXPublicKey()。
	publicKeyDer, err := g509.MarshalPKIXPublicKey(cert.certificate.PublicKey)
	if err != nil {
		return nil, err
	}

	cert.publicKey, err = g509.ParseSm2PublicKey(publicKeyDer)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// Sign
func (cert *ECDSACert) Sign(signBlock []byte) ([]byte, error) {
	return sm2Signature(cert.privateKey, signBlock)
}

// Verify
func (cert *ECDSACert) Verify(signBlock, signature []byte) error {
	return sm2Verify(cert.publicKey, signBlock, signature)
}

// EncryptAsn1 公钥加密
func (cert *ECDSACert) EncryptAsn1(plainBlock []byte) ([]byte, error) {
	return sm2Encrypt(cert.publicKey, plainBlock)
}

// DecryptAsn1 私钥解密
func (cert *ECDSACert) DecryptAsn1(cipherBlock []byte) ([]byte, error) {
	return sm2Decrypt(cert.privateKey, cipherBlock)
}

/*SerialNumber 获取证书序列号.*/
func (cert *ECDSACert) SerialNumber() string {
	return cert.certificate.SerialNumber.String()
}

/*SerialNumberHex 获取证书序列号十六进制（用于windows查看证书序列号为十六进制情况）.*/
func (cert *ECDSACert) SerialNumberHex() string {
	return string(EncodeHex(cert.certificate.SerialNumber.Bytes()))
}

/*Certificate 获取证书信息.*/
func (cert *ECDSACert) Certificate() *x509.Certificate {
	return cert.certificate.ToX509Certificate()
}

func decodePfxAllInfo(buf []byte, pass string) (*g509.Certificate, *sm2.PrivateKey, error) {
	pv, cer, err := p12.DecodeAll(buf, pass)
	if err != nil {
		return nil, nil, errors.Errorf("PFX证书加载失败[%s]", err)
	}

	switch k := pv.(type) {
	case *ecdsa.PrivateKey:
		switch k.Curve {
		case sm2.P256Sm2():
			sm2pub := &sm2.PublicKey{
				Curve: k.Curve,
				X:     k.X,
				Y:     k.Y,
			}
			sm2Pri := &sm2.PrivateKey{
				PublicKey: *sm2pub,
				D:         k.D,
			}
			if !k.IsOnCurve(k.X, k.Y) {
				return nil, nil, errors.New("error while validating SM2 private key: %v")
			}
			return cer[0], sm2Pri, nil
		}
	default:
		return nil, nil, errors.New("unexpected type for p12 private key")
	}

	return nil, nil, errors.New("key is nil")
}

func sm2Signature(privateKey *sm2.PrivateKey, content []byte) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.Errorf("privateKey is nil")
	}
	sign, err := privateKey.Sign(rand.Reader, content, nil)
	if err != nil {
		return nil, err
	}

	return sign, nil
}

func sm2Verify(publicKey *sm2.PublicKey, msg []byte, sign []byte) error {
	if publicKey == nil {
		return errors.Errorf("publicKey is nil")
	}

	ok := publicKey.Verify(msg, sign)

	if !ok {
		return errors.Errorf("Verify failed")
	} else {
		return nil
	}
}

func sm2Encrypt(publicKey *sm2.PublicKey, dataByte []byte) ([]byte, error) {
	if publicKey == nil {
		return nil, errors.Errorf("publicKey is nil")
	}

	cipherByte, err := publicKey.EncryptAsn1(dataByte, rand.Reader)
	if err != nil {
		return nil, err
	}

	return cipherByte, nil
}

func sm2Decrypt(privateKey *sm2.PrivateKey, cipherByte []byte) ([]byte, error) {
	if privateKey == nil {
		return nil, errors.Errorf("privateKey is nil")
	}
	dataByte, err := privateKey.DecryptAsn1(cipherByte)
	if err != nil {
		return nil, err
	}

	return dataByte, nil
}
