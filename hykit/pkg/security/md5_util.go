package security

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func Md5Base64(origData []byte) string {
	md5Ctx := md5.New()
	md5Ctx.Write(origData)
	return base64.StdEncoding.EncodeToString(md5Ctx.Sum(nil))
}

func Md5Hex(origData []byte) string {
	md5Ctx := md5.New()
	md5Ctx.Write(origData)
	return strings.ToUpper(hex.EncodeToString(md5Ctx.Sum(nil)))
}
func VerifyMd5Base64(origData []byte, desKey string) bool {
	md5Ctx := md5.New()
	md5Ctx.Write(origData)
	return base64.StdEncoding.EncodeToString(md5Ctx.Sum(nil)) == desKey
}

func Md5Byte(origData []byte) []byte {
	md5Ctx := md5.New()
	md5Ctx.Write(origData)
	return md5Ctx.Sum(nil)
}

func GetFileMd5(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", (Md5Byte(b))), nil
}

func HMacMd5(origData, key []byte) []byte {
	h := hmac.New(md5.New, key)
	h.Write(origData)
	return h.Sum(nil)
}
