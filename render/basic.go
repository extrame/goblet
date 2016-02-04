package render

import (
	"github.com/valyala/fasthttp"
	"html/template"
)

type RenderContext interface {
	StatusCode() int
	Layout() string
	Method() string
	TemplatePath() string
	BlockOptionType() string
	Callback() string
	Suffix() string
	Format() string
	//return the charset of the files
	CharSet() string
}

type RenderServer interface {
	WwwRoot() string
	PublicDir() string
	Env() string
}

//每一类的Render都必须返回一个RenderInstance用于具体的渲染
type RenderInstance interface {
	Render(ctx *fasthttp.RequestCtx, data interface{}, status int, funcs template.FuncMap) error
}

type Render interface {
	//返回一个RenderInstance用于具体的渲染
	PrepareInstance(RenderContext) (RenderInstance, error)
	//初始化
	Init(RenderServer, template.FuncMap)
}
