package security

import (
	"code.jshyjdtech.com/godev/hykit/pkg/security/ecb"
	"crypto/cipher"
	"fmt"
	"github.com/tjfoc/gmsm/sm4"
)

type SM4MOD int

const (
	SM4_CBC_PKCS5PADDING SM4MOD = iota
	SM4_CBC_PKCS7PADDING
	SM4_CBC_ZEROPADDING
	SM4_ECB_PKCS5PADDING
	SM4_ECB_PKCS7PADDING
	SM4_ECB_ZEROPADDING
)

func SM4EncryptIV(origData, key, iv []byte, mod SM4MOD) ([]byte, error) {
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = getPaddingSM4(mod, origData, blockSize)
	blockModee := cipher.NewCBCEncrypter(block, iv[:blockSize])
	crypted := make([]byte, len(origData))
	blockModee.CryptBlocks(crypted, origData)
	return crypted, nil
}

func SM4DecryptIV(crypted, key, iv []byte, mod SM4MOD) ([]byte, error) {
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockModee := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(crypted))
	blockModee.CryptBlocks(origData, crypted)
	return getUnpaddingSM4(mod, origData)
}

func SM4Encrypt(origData, key []byte, mod SM4MOD) ([]byte, error) {

	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = getPaddingSM4(mod, origData, blockSize)
	var blockMode cipher.BlockMode
	switch mod {
	case SM4_CBC_PKCS5PADDING, SM4_CBC_PKCS7PADDING, SM4_CBC_ZEROPADDING:
		blockMode = cipher.NewCBCEncrypter(block, key[:blockSize])
	case SM4_ECB_PKCS5PADDING, SM4_ECB_PKCS7PADDING, SM4_ECB_ZEROPADDING:
		blockMode = ecb.NewECBEncrypter(block)
	}
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func SM4Decrypt(crypted, key []byte, mod SM4MOD) ([]byte, error) {
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	var blockMode cipher.BlockMode
	switch mod {
	case SM4_CBC_PKCS5PADDING, SM4_CBC_PKCS7PADDING, SM4_CBC_ZEROPADDING:
		blockMode = cipher.NewCBCDecrypter(block, key[:blockSize])
	case SM4_ECB_PKCS5PADDING, SM4_ECB_PKCS7PADDING, SM4_ECB_ZEROPADDING:
		blockMode = ecb.NewECBDecrypter(block)
	}
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	return getUnpaddingSM4(mod, origData)
}

func getPaddingSM4(mod SM4MOD, data []byte, b_size int) []byte {
	switch mod {
	case SM4_CBC_PKCS5PADDING, SM4_ECB_PKCS5PADDING:
		return pPKCS5Padding(data, b_size)
	case SM4_CBC_PKCS7PADDING, SM4_ECB_PKCS7PADDING:
		return pPKCS7Padding(data, b_size)
	case SM4_CBC_ZEROPADDING, SM4_ECB_ZEROPADDING:
		return pZeroPadding(data, b_size)
	default:
		return nil
	}

}

func getUnpaddingSM4(mod SM4MOD, data []byte) ([]byte, error) {
	switch mod {
	case SM4_CBC_PKCS5PADDING, SM4_ECB_PKCS5PADDING:
		return pPKCS5UnPadding(data)
	case SM4_CBC_PKCS7PADDING, SM4_ECB_PKCS7PADDING:
		return pPKCS7UnPadding(data)
	case SM4_CBC_ZEROPADDING, SM4_ECB_ZEROPADDING:
		return pZeroUnPadding(data)
	default:
		return nil, fmt.Errorf("invalid PADDING MOD[%d]", mod)
	}
}
