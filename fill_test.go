package goblet

import (
	"fmt"
	"regexp"
	"testing"
)

func TestFillByMatch(t *testing.T) {
	var fillby_valid = regexp.MustCompile(`^\s*fillby\(\s*(\w*)\s*\)\s*$`)
	matched := fillby_valid.FindStringSubmatch("fillby(now)")
	if len(matched) != 2 || matched[1] != "now" {
		t.Fail()
	}
	matched = fillby_valid.FindStringSubmatch(" fillby(now)")
	if len(matched) != 2 || matched[1] != "now" {
		t.Errorf("fail in prefix empty")
	}
	matched = fillby_valid.FindStringSubmatch(" fillby(now) ")
	if len(matched) != 2 || matched[1] != "now" {
		t.Errorf("fail in both empty")
	}
	matched = fillby_valid.FindStringSubmatch(" fillby( now) ")
	if len(matched) != 2 || matched[1] != "now" {
		t.Errorf("fail in prefix empty in func")
	}
	matched = fillby_valid.FindStringSubmatch(" fillby(now ) ")
	if len(matched) != 2 || matched[1] != "now" {
		t.Errorf("fail in suffix empty in func")
	}
	matched = fillby_valid.FindStringSubmatch(" fillby( now ) ")
	if len(matched) != 2 || matched[1] != "now" {
		t.Errorf("fail in both empty in func")
	}
}

type A struct {
	Item string `goblet:"item"`
}

type B A

func TestAlias(t *testing.T) {
	form := map[string][]string{"item": []string{"value"}}
	b := new(B)
	UnmarshalForm(&form, b, true)
	fmt.Println(b)
}
