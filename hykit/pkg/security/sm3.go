package security

import (
	"fmt"
	"github.com/tjfoc/gmsm/sm3"
)

func SM3Encrypt(data string) []byte {
	h := sm3.New()
	h.Write([]byte(data))
	sum := h.Sum(nil)
	return sum
}

func SM3EncryptUpperString(data string) string {
	sum := SM3Encrypt(data)
	return fmt.Sprintf("%X", sum)
}
