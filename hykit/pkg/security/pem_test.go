package security

import (
	"encoding/hex"
	"io/ioutil"
	"testing"
)

func TestDecodePEM(t *testing.T) {
	body, err := ioutil.ReadFile("./data/pub.pem")
	if err != nil {
		t.Errorf("读取密钥文件失败")
		return
	}

	p, err := DecodePEM(body)
	if err != nil {
		t.Errorf("解析失败[%s]", err)
	}

	t.Logf("%+v", p)

}

func TestLoadCertPrivatePFX(t *testing.T) {
	body, err := ioutil.ReadFile("./data/unionpay_49982320.pfx")
	if err != nil {
		t.Errorf("读取密钥文件失败")
		return
	}

	c, err := LoadCertPrivatePFX(body, "000000")
	if err != nil {
		t.Errorf("加载证书失败[%s]", err)
		return
	}

	t.Logf("%s", c.SerialNumber())
	cert := c.Certificate()
	t.Logf("%s,%s,%s", cert.Subject, cert.SerialNumber.String(), hex.EncodeToString(cert.SerialNumber.Bytes()))

}

func TestLoadCertPrivateECPFX(t *testing.T) {
	body, err := ioutil.ReadFile("D:\\key\\SM2\\cert.pfx")
	if err != nil {
		t.Errorf("读取密钥文件失败")
		return
	}

	c, err := LoadECDSACertPrivatePFX(body, "123456")
	if err != nil {
		t.Errorf("加载证书失败[%s]", err)
		return
	}

	t.Logf("SerialNumber：%s", c.SerialNumber())
	cert := c.Certificate()
	t.Logf("Subject：%s, SerialNumber：%s, SerialNumber： %s", cert.Subject, cert.SerialNumber.String(), hex.EncodeToString(cert.SerialNumber.Bytes()))

}

func TestLoadCertPublicECDer(t *testing.T) {
	body, err := ioutil.ReadFile("D:\\key\\SM2\\cert.cer")
	if err != nil {
		t.Errorf("读取密钥文件失败")
		return
	}

	c, err := LoadECDSACertPublicCER(body)
	if err != nil {
		t.Errorf("加载证书失败[%s]", err)
		return
	}

	t.Logf("SerialNumber：%s", c.SerialNumber())
	cert := c.Certificate()
	t.Logf("Subject：%s, SerialNumber：%s, SerialNumber： %s", cert.Subject, cert.SerialNumber.String(), hex.EncodeToString(cert.SerialNumber.Bytes()))

}

func TestSM4Decrypt(t *testing.T) {
	plain := []byte("abc")
	key := []byte("A02C611E34561563")
	cipher, err := SM4Encrypt(plain, key, SM4_ECB_PKCS7PADDING)
	if err != nil {
		t.Logf("SM4加密失败[%s]", err)
		return
	}
	t.Logf("base64 cipher: %s", EncodeBase64(cipher))

	result, err := SM4Decrypt(cipher, key, SM4_ECB_PKCS7PADDING)
	if err != nil {
		t.Logf("SM4解密失败[%s]", err)
		return
	}
	t.Logf("result: %s", result)
}
