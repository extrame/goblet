package render

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"html/template"
)

type JsonRender struct {
}

func (j *JsonRender) PrepareInstance(c RenderContext) (RenderInstance, error) {
	if cb := c.Callback(); cb != "" {
		return &JsonCbRenderInstance{Cb: cb}, nil
	}
	return new(JsonRenderInstance), nil
}

func (j *JsonRender) Init(s RenderServer, funcs template.FuncMap) {
}

type JsonRenderInstance int8

func (r *JsonRenderInstance) Render(ctx *fasthttp.RequestCtx, data interface{}, status int, funcs template.FuncMap) (err error) {
	var v []byte
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	v, err = json.Marshal(data)
	if err == nil {
		ctx.Write(v)
	}
	return
}

type JsonCbRenderInstance struct {
	Cb string
}

func (r *JsonCbRenderInstance) Render(ctx *fasthttp.RequestCtx, data interface{}, status int, funcs template.FuncMap) (err error) {
	var v []byte
	ctx.SetStatusCode(status)
	v, err = json.Marshal(data)
	if err == nil {
		ctx.Write([]byte(r.Cb))
		ctx.Write([]byte("("))
		ctx.Write(v)
		ctx.Write([]byte(")"))
	}
	return
}
