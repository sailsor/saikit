package security

import (
	"encoding/base64"
	"encoding/hex"
)

func EncodeBase64(origdata []byte) []byte {
	str := base64.StdEncoding.EncodeToString(origdata)
	return []byte(str)
}

func DecodeBase64(cipherdata []byte) []byte {
	orig, err := base64.StdEncoding.DecodeString(string(cipherdata))
	if err != nil {
		return nil
	}
	return orig
}

func EncodeHex(origdata []byte) []byte {
	str := hex.EncodeToString(origdata)
	return []byte(str)
}

func DecodeHex(hexdata []byte) []byte {
	data, err := hex.DecodeString(string(hexdata))
	if err != nil {
		return nil
	}
	return data
}
