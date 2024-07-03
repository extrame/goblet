package render

import (
	"encoding/xml"
	"html/template"
	"io"
)

type XmlRender struct {
}

func (j *XmlRender) Type() string {
	return "xml"
}

func (j *XmlRender) PrepareInstance(c RenderContext) (RenderInstance, error) {
	return new(XmlRenderInstance), nil
}

func (j *XmlRender) Init(s RenderServer, funcs template.FuncMap) {
}

type XmlRenderInstance int8

func (r *XmlRenderInstance) Render(wr io.Writer, hwr HeadWriter, data interface{}, status int, funcs template.FuncMap) (err error) {
	var v []byte
	hwr.Header().Add("Content-Type", "text/xml;charset=UTF-8")
	hwr.WriteHeader(status)
	v, err = xml.Marshal(data)
	if err == nil {
		wr.Write(v)
	}
	return
}
