package render

import (
	"net/http"
)

type RenderContext interface {
	StatusCode() int
	Layout() string
	Method() string
	TemplatePath() string
	BlockOptionType() string
	Callback() string
}

type RenderServer interface {
	WwwRoot() string
	Env() string
}

//每一类的Render都必须返回一个RenderInstance用于具体的渲染
type RenderInstance interface {
	Render(wr http.ResponseWriter, data interface{}, status int) error
}

type Render interface {
	//返回一个RenderInstance用于具体的渲染
	PrepareInstance(RenderContext) (RenderInstance, error)
	//初始化
	Init(RenderServer)
}
