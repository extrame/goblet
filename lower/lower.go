package lower

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/valyala/fasthttp"
)

//Request 下层Request实现
type Context interface {
	ReqReferer() string
	ReqMethod() string
	ReqHeader() Header
	FormValue(string) string
	QueryValue(string) string
	RemoteAddr() string
	Body() io.Reader
	URL() *url.URL
}

type Writer interface{}

type Header interface{}

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
