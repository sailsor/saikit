package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

func IsCurrentDate(date string) bool {
	d := time.Now().Format("20060102")
	if d == date {
		return true
	}

	return false
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

func Signbuf(val interface{}) string {
	nameAdrr := make([]string, 0)
	keyVal := make(map[string]interface{}, 0)

	vt := reflect.TypeOf(val).Elem()
	va := reflect.ValueOf(val).Elem()
	for i := 0; i < va.NumField(); i++ {
		name := strings.Split(vt.Field(i).Tag.Get("json"), ",")[0]
		switch va.Field(i).Kind() {
		case reflect.Int64, reflect.Int:
			if value := va.Field(i).Int(); value != 0 {
				nameAdrr = append(nameAdrr, name)
				keyVal[name] = value
			}
		case reflect.String:
			if value := va.Field(i).String(); len(value) > 0 && name != "sign" {
				nameAdrr = append(nameAdrr, name)
				keyVal[name] = value
			}
		case reflect.Ptr:
			if !va.Field(i).IsNil() {
				nameAdrr = append(nameAdrr, name)
				keyVal[name] = va.Field(i).Elem()
			}
		}
	}
	sort.Strings(nameAdrr)
	var signBuf bytes.Buffer
	for _, name := range nameAdrr {
		signBuf.WriteString(fmt.Sprintf("%s=%v&", name, keyVal[name]))
	}
	signBuf.Truncate(signBuf.Len() - 1)
	return signBuf.String()
}

func Parse(resp string) (string, string) {
	index := strings.LastIndex(resp, "}{")
	if len(resp) >= index+1 {
		return resp[:index+1], resp[index+1:]
	}

	return "", ""
}

func Verifybuf(verifymsg []byte) string {
	nameAdrr := make([]string, 0)
	keyVal := make(map[string]interface{}, 0)

	json.Unmarshal(verifymsg, &keyVal)
	for name, value := range keyVal {
		nameAdrr = append(nameAdrr, name)
		switch value.(type) {
		case string:
			keyVal[name] = strings.TrimSpace(value.(string))
		case float64:
			keyVal[name] = fmt.Sprintf("%.0f", value.(float64))
		case int:
			keyVal[name] = fmt.Sprintf("%.0d", value.(int))
		case int64:
			keyVal[name] = fmt.Sprintf("%.0d", value.(int64))
		default:
			keyVal[name] = value
		}
	}

	sort.Strings(nameAdrr)
	var signBuf bytes.Buffer
	for _, name := range nameAdrr {
		signBuf.WriteString(fmt.Sprintf("%s=%v&", name, keyVal[name]))
	}
	signBuf.Truncate(signBuf.Len() - 1)
	return signBuf.String()
}
