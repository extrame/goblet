package render

import (
	"encoding/json"
	"net/http"
)

type JsonRender struct {
}

func (j *JsonRender) PrepareInstance(c RenderContext) (RenderInstance, error) {
	return new(JsonRenderInstance), nil
}

func (j *JsonRender) Init(s RenderServer) {

}

type JsonRenderInstance int8

func (r *JsonRenderInstance) Render(wr http.ResponseWriter, data interface{}, status int) (err error) {
	var v []byte
	wr.WriteHeader(status)
	v, err = json.Marshal(data)
	if err == nil {
		wr.Write(v)
	}
	return
}
