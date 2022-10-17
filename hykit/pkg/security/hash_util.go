package security

import (
	"crypto"
	"crypto/x509"
	"fmt"
)

const (
	// 证书RSA
	FILE_CERT_CER = "FILE_CERT_CER"
	FILE_CERT_PFX = "FILE_CERT_PFX"
	// 公私钥
	FILE_RSA_PEM_PUB    = "PEM-RSA-PUB"
	FILE_RSA_PEM_PRIV   = "PEM-RSA-PRIV"
	FILE_RSA_PKCS8_PRIV = "FILE_RSA_PKCS8_PRIV"
)

const (
	UnknownSignatureAlgorithm int = iota
	MD2WithRSA
	MD5WithRSA
	SHA1WithRSA
	SHA256WithRSA
	SHA384WithRSA
	SHA512WithRSA
	DSAWithSHA1
	DSAWithSHA256
	ECDSAWithSHA1
	ECDSAWithSHA256
	ECDSAWithSHA384
	ECDSAWithSHA512
)

func Hash(algo int, hashbuf []byte) ([]byte, error) {
	var hashType crypto.Hash

	switch algo {
	case SHA1WithRSA, DSAWithSHA1, ECDSAWithSHA1:
		hashType = crypto.SHA1
	case SHA256WithRSA, DSAWithSHA256, ECDSAWithSHA256:
		hashType = crypto.SHA256
	case SHA384WithRSA, ECDSAWithSHA384:
		hashType = crypto.SHA384
	case SHA512WithRSA, ECDSAWithSHA512:
		hashType = crypto.SHA512
	case MD2WithRSA, MD5WithRSA:
		hashType = crypto.MD5
	default:
		return nil, fmt.Errorf("非法的签名算法:[%d]", algo)
	}

	if !hashType.Available() {
		return nil, x509.ErrUnsupportedAlgorithm
	}
	h := hashType.New()
	h.Write(hashbuf)
	digest := h.Sum(nil)

	return digest, nil
}
