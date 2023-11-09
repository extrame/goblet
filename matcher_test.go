package goblet

import (
	"fmt"
	"testing"
)

type TestController struct {
	Route `/api/test`
	SingleController
}

type TestController2 struct {
	Route `/api`
	GroupController
}

func TestMatchSimpleController(t *testing.T) {
	matched, suffix := TestMatcher("/api/test/123/456", &TestController2{}, &TestController{})
	if suffix != "/123/456" {
		t.Error("suffix should be /123/456, but got", suffix)
	}
	if matched != "Html(TestController)" {
		t.Error("matched should be Html(TestController), but got", matched)
	}
}

func TestAttrMap(t *testing.T) {
	var lctx = &LoginContext{}
	WithAttribute("test", []string{"test1"})(lctx)
	if !lctx.HasAttr("test", "test1") {
		t.Error("should has attr test1")
	}
}

func TestCompareFloatAndInt(t *testing.T) {
	var result = compareStringAndNumber(float64(1), 1)
	fmt.Println("result is", result)
}
