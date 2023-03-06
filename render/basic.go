package render

import (
	"html/template"
	"io"
	"net/http"
)

type RenderContext interface {
	StatusCode() int
	Layout() string
	Method() string
	TemplatePath() string
	BlockOptionType() string
	Callback() string
	Format() string
	EnableCache()
	Version() string
	UseStandErrPage() bool
	UserAgent() string
	String() string
}

type RenderServer interface {
	WwwRoot() string
	PublicDir() string
	Env() string
	GetDelims() []string
}

//每一类的Render都必须返回一个RenderInstance用于具体的渲染
type RenderInstance interface {
	Render(wr io.Writer, header_wr HeadWriter, data interface{}, status int, funcs template.FuncMap) error
}

type HeadRenderInstance interface {
	HeadRender(wr io.Writer, header_wr HeadWriter, data interface{}, status int, funcs template.FuncMap) error
}

type HeadWriter interface {
	Header() http.Header
	WriteHeader(statusCode int)
}

type Render interface {
	//返回一个RenderInstance用于具体的渲染
	PrepareInstance(RenderContext) (RenderInstance, error)
	//初始化
	Init(RenderServer, template.FuncMap)
}
