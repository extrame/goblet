package render

import (
	"html/template"
	"net/http"
)

type RawRender int8

func (r *RawRender) PrepareInstance(c RenderContext) (RenderInstance, error) {
	return new(RawRenderInstance), nil
}

func (r *RawRender) Init(s RenderServer, funcs template.FuncMap) {
}

type RawRenderInstance int8

func (r *RawRenderInstance) Render(wr http.ResponseWriter, data interface{}, status int, funcs template.FuncMap) error {
	switch tdata := data.(type) {
	case []byte:
		wr.Write(tdata)
	}
	return nil
}
