package lower

import "github.com/valyala/fasthttp"

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
