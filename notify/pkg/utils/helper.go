package utils

import (
	"fmt"
	"net"
	"strings"
	"unicode"
)

func StringNotEmpty(args ...string) error {
	for _, v := range args {
		if strings.Trim(v, " ") == "" {
			return fmt.Errorf("字段[%v]不能为空", v)
		}
	}

	return nil
}

func TrimDupStringSlice(s []string) []string {
	if len(s) < 2 {
		return s
	}
	r := make([]string, 0)

	for i, v := range s {
		ok := false
		for _, v1 := range s[:i] {
			if v1 == v {
				ok = true
				break
			}
		}
		if !ok {
			r = append(r, v)
		}
	}
	return r
}

func RemoveSliceElem(src []string, rm []string) []string {
	var s []string

	for _, v := range src {
		if !StringInSlice(v, rm) {
			s = append(s, v)
		}
	}

	return s
}

func StringInSlice(s string, dts []string) bool {
	if len(dts) == 0 {
		return false
	}

	for _, v := range dts {
		if v == s {
			return true
		}
	}

	return false
}

func IntSliceConvert(slice []int64) string {
	return strings.Replace(strings.Trim(fmt.Sprint(slice), "[]"), " ", ",", -1)
}

func StringSliceConvert(slice []string) string {
	return "'" + strings.Join(slice, "','") + "'"
}

// SegmentSliceInt64 split int64 slice by step size
// [][]int64    the result of segmented int64 slice
func SegmentSliceInt64(elements []int64, step int) [][]int64 {
	l := len(elements)

	if step <= 0 {
		step = 10
	}

	n := l / step
	if t := l % step; t > 0 {
		n += 1
	}
	ret := make([][]int64, n)
	n--
	i := 0
	for i < n {
		if l := len(elements); l < step {
			break
		}
		ret[i] = elements[:step:step]
		elements = elements[step:]
		i++
	}
	ret[i] = elements
	return ret[:i+1]
}

// SegmentSliceString split string slice by step size
// [][]string    the result of segmented string slice
func SegmentSliceString(elements []string, step int) [][]string {
	l := len(elements)

	if step <= 0 {
		step = 10
	}

	n := l / step
	if t := l % step; t > 0 {
		n += 1
	}
	ret := make([][]string, n)
	n--
	i := 0
	for i < n {
		if l := len(elements); l < step {
			break
		}
		ret[i] = elements[:step:step]
		elements = elements[step:]
		i++
	}
	ret[i] = elements
	return ret[:i+1]
}

func GetLocalIp() (string, error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String(), nil
					}
				}
			}
		}
	}

	return "", err
}

// StringIsPrint 确认buff是否可打印 最多判断20个字符
func StringIsPrint(s string) bool {
	l := len(s)
	if l > 20 {
		l = 20
	}
	for i := 0; i < l; i++ {
		if unicode.IsPrint(rune(s[i])) {
			continue
		}
		return false
	}
	return true
}
