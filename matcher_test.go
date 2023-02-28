package goblet

import (
	"testing"
)

type TestController struct {
	Route `/api/test`
	SingleController
}

func TestMatchSimpleController(t *testing.T) {
	suffix, matched := TestMatcher(&TestController{}, "/api/test/123/456")
	if suffix != "/123/456" {
		t.Error("suffix should be /123/456, but got", suffix)
	}
	if !matched {
		t.Error("should be matched")
	}
}
