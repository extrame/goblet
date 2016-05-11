package lower

import (
	"github.com/valyala/fasthttp"
)

type fasthttpCtx struct {
	ctx *fasthttp.RequestCtx
}

func (f *fasthttpCtx) Referer() string {
	return string(f.ctx.Referer())
}

func (f *fasthttpCtx) Method() string {
	return string(f.ctx.Method())
}
