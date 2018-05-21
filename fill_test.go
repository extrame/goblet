package goblet

import (
	"fmt"
	"strings"
	"testing"
)

// func TestFillByMatch(t *testing.T) {
// 	var fillby_valid = regexp.MustCompile(`^\s*fillby\(\s*(\w*)\s*\)\s*$`)
// 	matched := fillby_valid.FindStringSubmatch("fillby(now)")
// 	if len(matched) != 2 || matched[1] != "now" {
// 		t.Fail()
// 	}
// 	matched = fillby_valid.FindStringSubmatch(" fillby(now)")
// 	if len(matched) != 2 || matched[1] != "now" {
// 		t.Errorf("fail in prefix empty")
// 	}
// 	matched = fillby_valid.FindStringSubmatch(" fillby(now) ")
// 	if len(matched) != 2 || matched[1] != "now" {
// 		t.Errorf("fail in both empty")
// 	}
// 	matched = fillby_valid.FindStringSubmatch(" fillby( now) ")
// 	if len(matched) != 2 || matched[1] != "now" {
// 		t.Errorf("fail in prefix empty in func")
// 	}
// 	matched = fillby_valid.FindStringSubmatch(" fillby(now ) ")
// 	if len(matched) != 2 || matched[1] != "now" {
// 		t.Errorf("fail in suffix empty in func")
// 	}
// 	matched = fillby_valid.FindStringSubmatch(" fillby( now ) ")
// 	if len(matched) != 2 || matched[1] != "now" {
// 		t.Errorf("fail in both empty in func")
// 	}
// }

// type A struct {
// 	Item string `goblet:"item,valu3"`
// }

// type B A

// func TestAlias(t *testing.T) {
// 	b := new(B)
// 	UnmarshalForm(func(tag string) []string {
// 		form := map[string][]string{"item": []string{"value"}}
// 		return form[tag]
// 	}, nil, b, true)
// 	fmt.Println(b)
// }

// func TestDefault(t *testing.T) {
// 	b := new(A)
// 	UnmarshalForm(func(tag string) []string {
// 		return []string{}
// 	}, nil, b, true)
// 	fmt.Println(b)
// }

// type C struct {
// 	Item string
// }

// func TestEmptyTag(t *testing.T) {
// 	b := new(C)
// 	UnmarshalForm(func(tag string) []string {
// 		fmt.Println(tag)
// 		form := map[string][]string{"Item": []string{"value"}}
// 		return form[tag]
// 	}, nil, b, true)
// 	fmt.Println(b)
// }

// type D struct {
// 	ExpectAt time.Time `goblet:",fillby(2006年1月2日)"`
// }

// func TestTimeFormat(t *testing.T) {
// 	b := new(D)
// 	UnmarshalForm(func(tag string) []string {
// 		fmt.Println(tag)
// 		form := map[string][]string{"ExpectAt": []string{"2015年11月2日"}}
// 		return form[tag]
// 	}, nil, b, true)
// 	fmt.Println(b)
// }

// func TestPkg(t *testing.T) {
// 	var file File
// 	r := reflect.ValueOf(file)
// 	ty := r.Type()
// 	log.Println(ty.PkgPath(), ty.Name())
// }

func TestBO(t *testing.T) {
	suffix := "123/10"
	args := strings.SplitN(suffix, "/", 2)
	fmt.Println(len(args), args[0])
}
