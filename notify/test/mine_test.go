package test

import (
	"mime"
	"testing"
)

func TestMime(t *testing.T) {
	// Content-Type:[application/x-www-form-urlencoded;charset:utf-8;]

	ct1 := "application/x-www-form-urlencoded;charset=utf-8;"
	t.Logf(ct1)
	ct1, _, err := mime.ParseMediaType(ct1)
	if err != nil {
		t.Logf("ct1 err:%s", err)
	} else {
		t.Logf("ct1 pass")
	}

	t.Logf("")

	ct2 := "application/x-www-form-urlencoded;"
	t.Logf(ct2)
	ct2, _, err = mime.ParseMediaType(ct2)
	if err != nil {
		t.Logf("ct2 err:%s", err)
	} else {
		t.Logf("ct2 pass")
	}

}

func TestMimeJson(t *testing.T) {
	// Content-Type:[application/json;charset:utf-8;]

	ct1 := "application/json;charset;utf-8;"
	t.Logf(ct1)
	ct1, _, err := mime.ParseMediaType(ct1)
	if err != nil {
		t.Logf("ct1 err:%s", err)
	} else {
		t.Logf("ct1 pass")
	}

	t.Logf("")

	ct2 := "application/json;charset=utf-8;"
	t.Logf(ct2)
	ct2, _, err = mime.ParseMediaType(ct2)
	if err != nil {
		t.Logf("ct2 err:%s", err)
	} else {
		t.Logf("ct2 pass")
	}

}
