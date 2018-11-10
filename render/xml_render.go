package render

import (
	"encoding/xml"
	"html/template"
	"net/http"
)

type XmlRender struct {
}

func (j *XmlRender) PrepareInstance(c RenderContext) (RenderInstance, error) {
	if cb := c.Callback(); cb != "" {
		return &XmlCbRenderInstance{Cb: cb}, nil
	}
	return new(XmlRenderInstance), nil
}

func (j *XmlRender) Init(s RenderServer, funcs template.FuncMap) {
}

type XmlRenderInstance int8

func (r *XmlRenderInstance) Render(wr http.ResponseWriter, data interface{}, status int, funcs template.FuncMap) (err error) {
	var v []byte
	wr.Header().Add("Content-Type", "application/Xml")
	wr.WriteHeader(status)
	v, err = xml.Marshal(data)
	if err == nil {
		wr.Write(v)
	}
	return
}

type XmlCbRenderInstance struct {
	Cb string
}

func (r *XmlCbRenderInstance) Render(wr http.ResponseWriter, data interface{}, status int, funcs template.FuncMap) (err error) {
	var v []byte
	wr.WriteHeader(status)
	v, err = xml.Marshal(data)
	if err == nil {
		wr.Write([]byte(r.Cb))
		wr.Write([]byte("("))
		wr.Write(v)
		wr.Write([]byte(")"))
	}
	return
}
