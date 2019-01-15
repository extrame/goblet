package render

import (
	"html/template"
	"io"
	"strconv"
)

type RawRender int8

func (r *RawRender) PrepareInstance(c RenderContext) (RenderInstance, error) {
	return new(RawRenderInstance), nil
}

func (r *RawRender) Init(s RenderServer, funcs template.FuncMap) {
}

type RawRenderInstance int8

func (r *RawRenderInstance) Render(wr io.Writer, hwr HeadWriter, data interface{}, status int, funcs template.FuncMap) (err error) {
	var writen = int64(0)
	switch tdata := data.(type) {
	case io.Reader:
		writen, err = io.Copy(wr, tdata)
	case []byte:
		var int_writen = 0
		int_writen, err = wr.Write(tdata)
		writen = int64(int_writen)
	}
	hwr.Header().Set("Content-Length", strconv.FormatInt(writen, 10))
	return err
}
