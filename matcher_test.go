package goblet

import (
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
