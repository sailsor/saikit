package security

import (
	"crypto/cipher"
	"crypto/des"
	"encoding/hex"
	"errors"
	"strings"
)

const gIV = "\x00\x00\x00\x00\x00\x00\x00\x00"

func GenAnsi99Mac(origData, key []byte) ([]byte, error) {
	var err error
	tmpR := make([]byte, 8)
	tmpI := make([]byte, 8)
	pData := pZeroPadding(origData, 8)
	l := len(pData)
	for i := 0; i < l; i += 8 {
		if i == 0 {
			copy(tmpR, gIV)
		}
		for j := 0; j < 8; j++ {
			tmpI[j] = tmpR[j] ^ pData[i+j]
		}
		tmpR, err = DesEncrypt(tmpI, key)
		if err != nil {
			return nil, err
		}
	}
	mac := strings.ToUpper(hex.EncodeToString(tmpR[:4]))
	return []byte(mac), nil
}

func DesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	origData = pZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, []byte(gIV))
	crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func DesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, []byte(gIV))
	origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	return origData, nil
}

// 3DES加密
func TripleDesEncryptZeroPD(origData, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	origData = pZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, []byte(gIV))
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// 3DES解密
func TripleDesDecryptZeroPD(crypted, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, []byte(gIV))
	origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	return getUnpadding(AES_CBC_ZEROPADDING, origData)
}

// 3DES加密
func TripleDesEncryptPKCS5PD(origData, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	origData = pPKCS5Padding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, []byte(gIV))
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// 3DES解密
func TripleDesDecryptPKCS5PD(crypted, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, []byte(gIV))
	origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	return getUnpadding(AES_CBC_PKCS5PADDING, origData)
}

// 3DES 16 ENC
func DoubleDesEncrypt(origData, key []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("非法的密钥长度")
	}
	tripKey := make([]byte, 0)
	tripKey = append(tripKey, key...)
	tripKey = append(tripKey, key[:8]...)

	block, err := des.NewTripleDESCipher(tripKey)
	if err != nil {
		return nil, err
	}
	origData = pZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, []byte(gIV))
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// 3DES 16 DEC
func DoubleDesDecrypt(cipherData, key []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("非法的密钥长度")
	}
	tripKey := make([]byte, 0)
	tripKey = append(tripKey, key...)
	tripKey = append(tripKey, key[:8]...)

	block, err := des.NewTripleDESCipher(tripKey)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, []byte(gIV))
	origData := cipherData
	blockMode.CryptBlocks(origData, cipherData)
	return getUnpadding(AES_CBC_PKCS5PADDING, origData)
}

// Des加密
func encrypt(origData, key []byte) ([]byte, error) {
	if len(origData) < 1 || len(key) < 1 {
		return nil, errors.New("wrong data or key")
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	if len(origData)%bs != 0 {
		return nil, errors.New("wrong padding")
	}
	out := make([]byte, len(origData))
	dst := out
	for len(origData) > 0 {
		block.Encrypt(dst, origData[:bs])
		origData = origData[bs:]
		dst = dst[bs:]
	}
	return out, nil
}

// decrypt Des解密
func decrypt(crypted, key []byte) ([]byte, error) {
	if len(crypted) < 1 || len(key) < 1 {
		return nil, errors.New("wrong data or key")
	}
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(crypted))
	dst := out
	bs := block.BlockSize()
	if len(crypted)%bs != 0 {
		return nil, errors.New("wrong crypted size")
	}

	for len(crypted) > 0 {
		block.Decrypt(dst, crypted[:bs])
		crypted = crypted[bs:]
		dst = dst[bs:]
	}

	return out, nil
}

func TripleDesEncryptEcb(origData, key []byte) ([]byte, error) {
	tkey := make([]byte, 24)
	copy(tkey, key)
	k1 := tkey[:8]
	k2 := tkey[8:16]
	k3 := tkey[16:]

	block, err := des.NewCipher(k1)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	origData = getPadding(AES_ECB_PKCS5PADDING, origData, bs)

	buf1, err := encrypt(origData, k1)
	if err != nil {
		return nil, err
	}
	buf2, err := decrypt(buf1, k2)
	if err != nil {
		return nil, err
	}
	out, err := encrypt(buf2, k3)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func TripleDesDecryptEcb(crypted, key []byte) ([]byte, error) {
	tripKey := make([]byte, 24)
	copy(tripKey, key)
	k1 := tripKey[:8]
	k2 := tripKey[8:16]
	k3 := tripKey[16:]
	buf1, err := decrypt(crypted, k3)
	if err != nil {
		return nil, err
	}
	buf2, err := encrypt(buf1, k2)
	if err != nil {
		return nil, err
	}
	out, err := decrypt(buf2, k1)
	if err != nil {
		return nil, err
	}
	return getUnpadding(AES_ECB_PKCS5PADDING, out)
}
