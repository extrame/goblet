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
	Route `/api/:id/test/:param`
	GroupController
}

func TestMatchSimpleController(t *testing.T) {
	matched, suffix, params := TestMatcher("/api/test/123/456", &TestController2{}, &TestController{})
	if suffix != "/123/456" {
		t.Error("suffix should be /123/456, but got", suffix)
	}
	if matched != "Html(TestController)" {
		t.Error("matched should be Html(TestController), but got", matched)
	}
	if params["id"] != "123" {
		t.Error("params[id] should be 123, but got", params["id"])
	}
	if params["param"] != "456" {
		t.Error("params[param] should be 456, but got", params["param"])
	}
}

func TestMatchParamController(t *testing.T) {
	matched, suffix, params := TestMatcher("/api/123/test/456", &TestController2{}, &TestController{})
	if suffix != "/123/456" {
		t.Error("suffix should be /123/456, but got", suffix)
	}
	if matched != "Html(TestController2)" {
		t.Error("matched should be Html(TestController2), but got", matched)
	}
	if params["id"] != "123" {
		t.Error("params[id] should be 123, but got", params["id"])
	}
	if params["param"] != "456" {
		t.Error("params[param] should be 456, but got", params["param"])
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
