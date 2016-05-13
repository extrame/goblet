package lower

import (
	"bytes"
	"io"
	"net/url"

	"github.com/valyala/fasthttp"
)

type fasthttpCtx struct {
	ctx *fasthttp.RequestCtx
}

func (f *fasthttpCtx) ReqReferer() string {
	return string(f.ctx.Referer())
}

func (f *fasthttpCtx) ReqMethod() string {
	return string(f.ctx.Method())
}

func (f *fasthttpCtx) FormValue(key string) string {
	bts := f.ctx.QueryArgs().Peek(key)
	if len(bts) == 0 {
		bts = f.ctx.PostArgs().Peek(key)
	}
	return string(bts)
}

func (f *fasthttpCtx) QueryValue(key string) string {
	return string(f.ctx.URI().QueryArgs().Peek(key))
}

func (f *fasthttpCtx) RemoteAddr() string {
	return f.ctx.RemoteAddr().String()
}

func (f *fasthttpCtx) Body() io.Reader {
	return bytes.NewBuffer(f.ctx.Request.Body())
}

func (f *fasthttpCtx) URL() *url.URL {
	u, _ := url.Parse(string(f.ctx.RequestURI()))
	return u
}

type fastURL struct {
	*fasthttp.URI
}
