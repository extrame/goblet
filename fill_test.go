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
	Item string `goblet:"item,valu3"`
}

type B A

func TestAlias(t *testing.T) {
	b := new(B)
	UnmarshalForm(func(tag string) []string {
		form := map[string][]string{"item": []string{"value"}}
		return form[tag]
	}, b, true)
	fmt.Println(b)
}

func TestDefault(t *testing.T) {
	b := new(A)
	UnmarshalForm(func(tag string) []string {
		return []string{}
	}, b, true)
	fmt.Println(b)
}
