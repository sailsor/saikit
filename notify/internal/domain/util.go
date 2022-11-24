package domain

import (
	"bytes"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"
)

var fieldName = func(v string) string {
	return strings.Split(v, ",")[0]
}

func SignatureBuffer(val interface{}) string {
	nameArr := make([]string, 0)
	keyVal := make(map[string]string, 0)
	v := reflect.ValueOf(val).Elem()
	st := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.CanInterface() {
			if f.Type().Kind() == reflect.Interface {
				continue
			}
			if f.Type().Kind() == reflect.Struct {
				sf := f.Type()
				for j := 0; j < sf.NumField(); j++ {
					name := fieldName(sf.Field(j).Tag.Get("form"))
					value := f.Field(j).String()
					if len(value) > 0 &&
						len(name) > 0 &&
						name != "signature" {
						nameArr = append(nameArr, name)
						keyVal[name] = value
					}
				}
				continue
			}
			name := fieldName(st.Field(i).Tag.Get("form"))
			value := f.String()
			if len(value) > 0 &&
				len(name) > 0 &&
				name != "signature" {
				nameArr = append(nameArr, name)
				keyVal[name] = value
			}
		}

	}
	sort.Strings(nameArr)
	var signBuf bytes.Buffer
	for _, name := range nameArr {
		signBuf.WriteString(fmt.Sprintf("%s=%s&", name, keyVal[name]))
	}
	signBuf.Truncate(signBuf.Len() - 1)
	return signBuf.String()
}

func Marshal(val interface{}) string {
	v := reflect.ValueOf(val).Elem()
	st := v.Type()
	u := url.Values{}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.CanInterface() {
			if f.Type().Kind() == reflect.Interface {
				continue
			}
			name := fieldName(st.Field(i).Tag.Get("form"))
			if name == "-" {
				continue
			}
			value := f.String()
			if len(value) > 0 {
				u.Add(name, value)
			}
		}
	}
	return u.Encode()
}
