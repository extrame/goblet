package plugin

import (
	"html/template"
	"sync"

	"github.com/extrame/goblet/v2"
	"github.com/extrame/goblet/v2/render"
)

type HtmlRenderPlugin struct {
	root      *template.Template
	dir       string
	models    *sync.Map
	pathRoot  *sync.Map
	suffix    string
	public    string
	saveTemp  bool
	notExists sync.Map
	delims    []string
}

func (h *HtmlRenderPlugin) Type() string {
	return "html"
}

func (h *HtmlRenderPlugin) PrepareInstance(c render.RenderContext) (render.RenderInstance, error) {
	// 实现与原来HtmlRender相同的PrepareInstance逻辑
	// 这里省略具体实现...
	return nil, nil
}

func (h *HtmlRenderPlugin) Init(s render.RenderServer, funcs template.FuncMap) {
	// 实现与原来HtmlRender相同的Init逻辑
	// 这里省略具体实现...
}

// 实现NewPlugin接口
func (h *HtmlRenderPlugin) AddCfgAndInit(s *goblet.Server) error {
	s.Renders[h.Type()] = h
	return nil
}
