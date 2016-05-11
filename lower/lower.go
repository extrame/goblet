package lower

import (
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"
)

//Request 下层Request实现
type Context interface {
	ReqReferer() string
	ReqMethod() string
	FormValue(string) string
	QueryValue(string) string
}

type Writer interface{}

func Wrap(base string, request interface{}, writer interface{}) (Context, error) {
	switch base {
	case "http":
		if r, ok := request.(*http.Request); ok {
			if w, ok := request.(*http.ResponseWriter); ok {
				return &netHTTPLowerReqeust{r, w}, nil
			}
		}
	case "fasthttp":
		if r, ok := request.(*fasthttp.RequestCtx); ok {
			ctx := fasthttpCtx{r}
			return &ctx, nil
		}
	}
	return nil, fmt.Errorf("input content is error")
}
