package plugin

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"

	"github.com/extrame/goblet"
	"github.com/extrame/goblet/render"
)

// StandardJsonWrapper return a standard json wrapper
// StandardJsonWrapper 返回一个标准的json包装器
// successCode: the code of success
// successMsg: the message of success
func StandardJsonWrapper(arguments ...StandardJsonWrapperSetter) render.Render {
	var wrapper = &jsonRenderWrapper{
		successCode: 200,
		successMsg:  "success",
	}
	for _, setter := range arguments {
		setter(wrapper)
	}
	return wrapper
}

// StandardJsonWrapperSetter set the success code and message
// StandardJsonWrapperSetter 设置成功的代码和消息
type StandardJsonWrapperSetter func(*jsonRenderWrapper)

func SjwWithSuccessCode(code int) StandardJsonWrapperSetter {
	return func(j *jsonRenderWrapper) {
		j.successCode = code
	}
}

func SjwWithSuccessMsg(msg string) StandardJsonWrapperSetter {
	return func(j *jsonRenderWrapper) {
		j.successMsg = msg
	}
}

type jsonRenderWrapper struct {
	successCode int
	successMsg  string
}

func (j *jsonRenderWrapper) Type() string {
	return "json"
}

func (j *jsonRenderWrapper) PrepareInstance(c render.RenderContext) (render.RenderInstance, error) {
	if cb := c.Callback(); cb != "" {
		return &jsonCbRenderInstance{Cb: cb}, nil
	}
	return new(jsonRenderInstance), nil
}

func (p *jsonRenderWrapper) RespendOk(ctx *goblet.Context) {
	ctx.Respond(nil)
}

func (p *jsonRenderWrapper) RespondError(ctx *goblet.Context, err error, context ...string) {
	var errCode = 500
	var standardData *StandardErrorOrData
	var ok bool
	standardData, ok = err.(*StandardErrorOrData)
	if !ok {
		standardData = &StandardErrorOrData{Data: nil, Msg: err.Error(), Code: errCode}
	}
	ctx.Respond(&standardData)
}

func (p *jsonRenderWrapper) DefaultRender() string {
	return "json"
}

func (j *jsonRenderWrapper) Init(s render.RenderServer, funcs template.FuncMap) {
}

type jsonRenderInstance int8

type StandardErrorOrData struct {
	Data interface{} `json:"data,omitempty"`
	Msg  string      `json:"msg"`
	Code int         `json:"code"`
}

func (s *StandardErrorOrData) Error() string {
	return fmt.Sprintf("Code: %d, Msg: %s", s.Code, s.Msg)
}

func (r *jsonRenderInstance) Render(wr io.Writer, hwr render.HeadWriter, data interface{}, status int, funcs template.FuncMap) (err error) {
	var v []byte
	hwr.Header().Add("Content-Type", "application/json; charset=utf-8")
	hwr.WriteHeader(status)
	if err, ok := data.(*StandardErrorOrData); !ok {
		data = StandardErrorOrData{Data: data, Msg: "success", Code: 0}
	} else {
		data = err
	}
	v, err = json.Marshal(&data)
	if err == nil {
		wr.Write(v)
	}
	return
}

type jsonCbRenderInstance struct {
	Cb string
}

func (r *jsonCbRenderInstance) Render(wr io.Writer, hwr render.HeadWriter, data interface{}, status int, funcs template.FuncMap) (err error) {
	var v []byte
	hwr.WriteHeader(status)
	if err, ok := data.(*StandardErrorOrData); !ok {
		data = StandardErrorOrData{Data: data, Msg: "success", Code: 0}
	} else {
		data = err
	}
	v, err = json.Marshal(&data)
	if err == nil {
		wr.Write([]byte(r.Cb))
		wr.Write([]byte("("))
		wr.Write(v)
		wr.Write([]byte(")"))
	}
	return
}

func StdError(code int, msg string) error {
	return &StandardErrorOrData{Data: nil, Msg: msg, Code: code}
}

func WrapError(code int, err error) error {
	if err == nil {
		return nil
	}
	return &StandardErrorOrData{Data: nil, Msg: err.Error(), Code: code}
}
