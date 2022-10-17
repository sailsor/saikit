package security

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/pkg/errors"
)

type RSAKey struct {
	KeyType    string `json:"key_type"`
	KeyFile    string `json:"key_file"`
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
	Modules    int
}

func NewRSAInMap(cfg map[string]string) (*RSAKey, error) {
	r := new(RSAKey)

	r.KeyFile = cfg["KEY_FILE"]
	r.KeyType = cfg["KEY_TYPE"]

	keyBuf, err := ioutil.ReadFile(r.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("读取RSA密钥失败[%s]", r.KeyFile)
	}

	block, err := DecodePEM(keyBuf)
	if err != nil {
		return nil, errors.Errorf("读取RSA密钥文件失败[%s][%s]", r.KeyFile, err)
	}

	var ok bool
	switch r.KeyType {
	case FILE_RSA_PEM_PUB:
		var pub interface{}
		pub, err = x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("解析RSA公钥失败[%s]", err)
		}
		r.PublicKey, ok = pub.(*rsa.PublicKey)
		if !ok {
			return nil,
				fmt.Errorf("Value returned from ParsePKIXPublicKey was not an  public key")
		}
		r.Modules = publicKeyModules(r.PublicKey)
	case FILE_RSA_PEM_PRIV:
		r.PrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("ParsePKCS1PrivateKey err:%s", err)
		}
		r.PublicKey = &r.PrivateKey.PublicKey
		r.Modules = privateKeyModules(r.PrivateKey)
	case FILE_RSA_PKCS8_PRIV:
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("ParsePKCS1PrivateKey err:%s", err)
		}
		r.PrivateKey = key.(*rsa.PrivateKey)
		r.PublicKey = &r.PrivateKey.PublicKey
		r.Modules = privateKeyModules(r.PrivateKey)
	default:
		return nil, fmt.Errorf("非法的RSA公钥类型[%s]", r.KeyType)
	}
	return r, nil
}

func publicKeyModules(pub *rsa.PublicKey) int {
	return (pub.N.BitLen() + 7) / 8
}

func privateKeyModules(priv *rsa.PrivateKey) int {
	return (priv.N.BitLen() + 7) / 8
}

/*使用RSA公钥装载*/
func loadFromRSAPublic(pub *rsa.PublicKey) *RSAKey {
	r := new(RSAKey)
	r.PublicKey = pub
	r.Modules = publicKeyModules(r.PublicKey)
	return r
}

/*使用RSA私钥装载*/
func loadFromRSAPrivate(priv *rsa.PrivateKey) *RSAKey {
	r := new(RSAKey)
	r.PrivateKey = priv
	r.PublicKey = &r.PrivateKey.PublicKey
	r.Modules = privateKeyModules(r.PrivateKey)
	return r
}

/*使用RSA公钥装载*/
func LoadRSAPubicPEM(buf []byte) (*RSAKey, error) {
	r := new(RSAKey)

	pemBlock, err := DecodePEM(buf)
	if err != nil {
		return nil, errors.Errorf("读取PEM格式RSA公钥失败[%s][%s]", buf, err)
	}

	pub, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析RSA公钥失败[%s]", err)
	}
	var ok bool
	r.PublicKey, ok = pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Value returned from ParsePKIXPublicKey was not an RSA public key")
	}
	r.Modules = publicKeyModules(r.PublicKey)
	return r, nil
}

/*使用RSA私钥装载*/
func LoadRSAPrivatePEM(buf []byte) (*RSAKey, error) {
	r := new(RSAKey)
	var err error
	block, err := DecodePEM(buf)
	if err != nil {
		return nil, errors.Errorf("PEM Decode failed[%s] err[%s]", buf, err)
	}

	r.PrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("ParsePKCS1PrivateKey err:%s", err)
	}
	r.PublicKey = &r.PrivateKey.PublicKey
	r.Modules = privateKeyModules(r.PrivateKey)
	return r, nil
}

func (r *RSAKey) Sign(algo int, signBlock []byte) ([]byte, error) {
	var hashType crypto.Hash
	switch algo {
	case SHA1WithRSA:
		hashType = crypto.SHA1
	case SHA256WithRSA:
		hashType = crypto.SHA256
	case SHA384WithRSA:
		hashType = crypto.SHA384
	case SHA512WithRSA:
		hashType = crypto.SHA512
	default:
		return nil, x509.InsecureAlgorithmError(algo)
	}
	if !hashType.Available() {
		return nil, x509.ErrUnsupportedAlgorithm
	}
	h := hashType.New()
	h.Write(signBlock)
	digest := h.Sum(nil)

	return r.PrivateKey.Sign(rand.Reader, digest, hashType)
}

func (r *RSAKey) Verify(algo int, signbuf, signature []byte) error {
	var hashType crypto.Hash

	switch algo {
	case SHA1WithRSA:
		hashType = crypto.SHA1
	case SHA256WithRSA:
		hashType = crypto.SHA256
	case SHA384WithRSA:
		hashType = crypto.SHA384
	case SHA512WithRSA:
		hashType = crypto.SHA512
	default:
		return fmt.Errorf("非法的验签算法:[%d]", algo)
	}

	if !hashType.Available() {
		return x509.ErrUnsupportedAlgorithm
	}
	h := hashType.New()
	h.Write(signbuf)
	digest := h.Sum(nil)

	return rsa.VerifyPKCS1v15(r.PublicKey, hashType, digest, signature)
}

func (r *RSAKey) RSAPublicEncryptPKCS1v15(origBlock []byte) ([]byte, error) {
	encLen := r.Modules - 11 // PKCS1v15
	orig := bytes.NewBuffer(origBlock)

	var ciperdata bytes.Buffer

	for {
		data := orig.Next(encLen)
		if len(data) == 0 {
			break
		}
		cipher, err := rsa.EncryptPKCS1v15(rand.Reader, r.PublicKey, data)
		if err != nil {
			return nil, fmt.Errorf("rsa.EncryptPKCS1v15 error[%s]", err)
		}
		ciperdata.Write(cipher)
	}
	return ciperdata.Bytes(), nil
}

func (r *RSAKey) RSAPublicDecryptPKCS1v15(cipherdata []byte) ([]byte, error) {
	encLen := r.Modules // PKCS1v15
	orig := bytes.NewBuffer(cipherdata)

	var origBlock bytes.Buffer

	for {
		data := orig.Next(encLen)
		if len(data) == 0 {
			break
		}
		cipher := publicDecrypt(r.PublicKey, data)
		origBlock.Write(cipher)
	}
	return origBlock.Bytes(), nil
}

func publicDecrypt(pubKey *rsa.PublicKey, data []byte) []byte {
	c := new(big.Int)
	m := new(big.Int)
	m.SetBytes(data)
	e := big.NewInt(int64(pubKey.E))
	c.Exp(m, e, pubKey.N)
	out := c.Bytes()
	skip := 0
	for i := 2; i < len(out); i++ {
		if i+1 >= len(out) {
			break
		}
		if out[i] == 0xff && out[i+1] == 0 {
			skip = i + 2
			break
		}
	}
	return out[skip:]
}

func (r *RSAKey) RSAPrivateEncryptPKCS1v15(origBlock []byte) ([]byte, error) {
	encLen := r.Modules - 11 // PKCS1v15
	orig := bytes.NewBuffer(origBlock)

	var ciperdata bytes.Buffer

	for {
		data := orig.Next(encLen)
		if len(data) == 0 {
			break
		}
		cipher, err := rsa.SignPKCS1v15(rand.Reader, r.PrivateKey, crypto.Hash(0), data)
		if err != nil {
			return nil, fmt.Errorf("rsa.EncryptPKCS1v15 error[%s]", err)
		}
		ciperdata.Write(cipher)
	}
	return ciperdata.Bytes(), nil
}

func (r *RSAKey) RSAPrivateDecryptPKCS1v15(cipherdata []byte) ([]byte, error) {
	encLen := r.Modules // PKCS1v15
	orig := bytes.NewBuffer(cipherdata)

	var origBlock bytes.Buffer

	for {
		data := orig.Next(encLen)
		if len(data) == 0 {
			break
		}
		cipher, err := rsa.DecryptPKCS1v15(rand.Reader, r.PrivateKey, data)
		if err != nil {
			return nil, fmt.Errorf("rsa.DecryptPKCS1v15 error[%s]", err)
		}
		origBlock.Write(cipher)
	}
	return origBlock.Bytes(), nil
}

func (r *RSAKey) GetModules() string {
	return fmt.Sprintf("%d", r.Modules*8)
}
