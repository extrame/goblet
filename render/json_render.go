package render

import (
	"encoding/json"
	"html/template"
	"io"
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

func (r *JsonRenderInstance) Render(wr io.Writer, hwr HeadWriter, data interface{}, status int, funcs template.FuncMap) (err error) {
	var v []byte
	hwr.Header().Add("Content-Type", "application/json; charset=utf-8")
	hwr.WriteHeader(status)
	v, err = json.Marshal(data)
	if err == nil {
		wr.Write(v)
	}
	return
}

type JsonCbRenderInstance struct {
	Cb string
}

func (r *JsonCbRenderInstance) Render(wr io.Writer, hwr HeadWriter, data interface{}, status int, funcs template.FuncMap) (err error) {
	var v []byte
	hwr.WriteHeader(status)
	v, err = json.Marshal(data)
	if err == nil {
		wr.Write([]byte(r.Cb))
		wr.Write([]byte("("))
		wr.Write(v)
		wr.Write([]byte(")"))
	}
	return
}
