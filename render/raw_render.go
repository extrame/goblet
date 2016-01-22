package render

import (
	"github.com/valyala/fasthttp"
	"html/template"
)

type RawRender int8

func (r *RawRender) PrepareInstance(c RenderContext) (RenderInstance, error) {
	return new(RawRenderInstance), nil
}

func (r *RawRender) Init(s RenderServer, funcs template.FuncMap) {
}

type RawRenderInstance int8

func (r *RawRenderInstance) Render(ctx *fasthttp.RequestCtx, data interface{}, status int, funcs template.FuncMap) error {
	switch tdata := data.(type) {
	case []byte:
		ctx.Write(tdata)
	}
	return nil
}
