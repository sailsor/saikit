package sequence

import (
	"crypto/rand"
	"math/big"
	"strconv"
	"time"

	"github.com/sony/sonyflake"
)

var sf *sonyflake.Sonyflake

func init() {
	curr, _ := time.Parse("20060102", "20200101")
	//rand.Seed(time.Now().UnixNano())
	//offset := rand.Int() % 99
	a, _ := rand.Int(rand.Reader, big.NewInt(99))
	offset := a.Int64()
	curr = curr.Add(time.Duration(offset) * 24 * time.Hour)
	settings := sonyflake.Settings{
		StartTime: curr,
	}
	sf = sonyflake.NewSonyflake(settings)
}

func GenID() (string, error) {
	id, err := sf.NextID()
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(id, 10), nil
}
