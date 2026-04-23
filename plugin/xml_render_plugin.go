package plugin

import (
	"html/template"

	"github.com/extrame/goblet/v2"
	"github.com/extrame/goblet/v2/render"
)

type XmlRenderPlugin struct {
}

func (x *XmlRenderPlugin) Type() string {
	return "xml"
}

func (x *XmlRenderPlugin) PrepareInstance(c render.RenderContext) (render.RenderInstance, error) {
	return new(render.XmlRenderInstance), nil
}

func (x *XmlRenderPlugin) Init(s render.RenderServer, funcs template.FuncMap) {
}

// 实现NewPlugin接口
func (x *XmlRenderPlugin) AddCfgAndInit(s *goblet.Server) error {
	s.Renders[x.Type()] = x
	return nil
}
