package goblet

import (
	"reflect"
	"strings"
)

type Route byte
type Render byte
type Layout byte

const (
	REST_READ       = "read"
	REST_READMANY   = "readmany"
	REST_DELETE     = "delete"
	REST_NEW        = "new"
	REST_CREATE     = "create"
	REST_UPDATE     = "update"
	REST_UPDATEMANY = "updatemany"
)

type BlockOption interface {
	UpdateRender(string, *Context)
	GetRouting() []string
	MatchSuffix(string) bool
	Parse(*Context) error
	Layout() string
}

type BasicBlockOption struct {
	routing             []string
	render              []string
	layout              string
	htmlRenderFileOrDir string
	block               interface{}
}

type HtmlBlockOption struct {
	BasicBlockOption
}

func (h *HtmlBlockOption) MatchSuffix(suffix string) bool {
	return len(suffix) == 0
}

func (h *HtmlBlockOption) Parse(c *Context) error {
	if c.Request.Method == "GET" {
		if get, ok := h.BasicBlockOption.block.(HtmlGetBlock); ok {
			get.Get(c)
		}
	} else if c.Request.Method == "POST" {
		if post, ok := h.BasicBlockOption.block.(HtmlPostBlock); ok {
			post.Post(c)
		}
	}
	return nil
}

func (b *BasicBlockOption) Layout() string {
	return b.layout
}

func (h *BasicBlockOption) UpdateRender(o string, ctx *Context) {
	h.htmlRenderFileOrDir = o
}

func (b *BasicBlockOption) GetRouting() []string {
	return b.routing
}

type RestBlockOption struct {
	BasicBlockOption
}

func (r *RestBlockOption) UpdateRender(obj string, ctx *Context) {
	ctx.method = obj
}

func (r *RestBlockOption) Parse(c *Context) error {
	var method string
	if method = c.Request.URL.Query().Get("method"); method == "" {
		method = c.Request.Method
	}
	method = strings.ToUpper(method)
	if len(c.suffix) > 0 {
		id := c.suffix[1:]
		if id == "new" {
			r.renderAsNew(c)
		} else if method == "GET" {
			r.renderAsRead(id, c)
		}
	} else {
		if method == "GET" {
			r.renderAsReadMany(c)
		} else if method == "POST" {
			r.renderAsCreate(c)
		} else if method == "PUT" {
			r.renderAsUpdateMany(c)
		}
	}

	return nil
}

func (r *RestBlockOption) renderAsRead(id string, cx *Context) {
	if reader, ok := r.BasicBlockOption.block.(RestReadBlock); ok {
		cx.method = REST_READ
		reader.Read(id, cx)
	}
}

func (r *RestBlockOption) renderAsReadMany(cx *Context) {
	if reader, ok := r.BasicBlockOption.block.(RestReadManyBlock); ok {
		cx.method = REST_READMANY
		reader.ReadMany(cx)
	}
}

func (r *RestBlockOption) renderAsNew(cx *Context) {
	if reader, ok := r.BasicBlockOption.block.(RestNewBlock); ok {
		cx.method = REST_NEW
		reader.New(cx)
	}
}

func (r *RestBlockOption) renderAsUpdateMany(cx *Context) {
	if um, ok := r.BasicBlockOption.block.(RestUpdateManyBlock); ok {
		cx.method = REST_UPDATEMANY
		um.UpdateMany(cx)
	}
}

func (r *RestBlockOption) renderAsCreate(cx *Context) {
	if um, ok := r.BasicBlockOption.block.(RestCreateBlock); ok {
		cx.method = REST_CREATE
		um.Create(cx)
	}
}

func (r *RestBlockOption) handleData(c *Context) {

}

func (r *RestBlockOption) MatchSuffix(suffix string) bool {
	return len(suffix) == 0 || suffix[0:1] == "/"
}

type CommonBlokOption struct {
	BasicBlockOption
}

func (c *CommonBlokOption) MatchSuffix(suffix string) bool {
	return true
}

func (c *CommonBlokOption) Parse(*Context) error {
	return nil
}

func PrepareOption(block interface{}) BlockOption {
	opt := newBlock(block)
	var basic *BasicBlockOption
	switch typ := opt.(type) {
	case *HtmlBlockOption:
		basic = &typ.BasicBlockOption
	case *RestBlockOption:
		basic = &typ.BasicBlockOption
	case *CommonBlokOption:
		basic = &typ.BasicBlockOption
	}

	basic.block = block
	val := reflect.ValueOf(block)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	valtype := val.Type()

	if val.Kind() == reflect.Struct {
		for i := 0; i < valtype.NumField(); i++ {
			t := valtype.Field(i)
			// v := val.Field(i)
			tags := strings.Split(string(t.Tag), ",")
			if t.Type.Name() == "Route" && t.Type.PkgPath() == "github.com/extrame/goblet" {
				basic.routing = tags
				continue
			}
			if t.Type.Name() == "Render" && t.Type.PkgPath() == "github.com/extrame/goblet" {
				basic.render = make([]string, len(tags))
				for k, v := range tags {
					vs := strings.Split(v, "=")
					basic.render[k] = vs[0]
					if vs[0] == "html" && len(vs) >= 2 {
						basic.htmlRenderFileOrDir = vs[1]
					}
				}
				continue
			}
			if t.Type.Name() == "Layout" && t.Type.PkgPath() == "github.com/extrame/goblet" {
				basic.layout = string(t.Tag)
				continue
			}
		}
	}

	if len(basic.routing) == 0 {
		basic.routing = []string{"/" + valtype.Name()}
	}

	if len(basic.render) == 0 {
		basic.routing = []string{"json"}
	}

	if basic.htmlRenderFileOrDir == "" {
		basic.htmlRenderFileOrDir = valtype.Name()
	}

	if basic.layout == "" {
		basic.layout = "default"
	}

	return opt
}

func newBlock(block interface{}) BlockOption {
	if _, ok := block.(HtmlGetBlock); ok {
		return &HtmlBlockOption{}
	} else if _, ok := block.(HtmlPostBlock); ok {
		return &HtmlBlockOption{}
	}

	if _, ok := block.(RestNewBlock); ok {
		return &RestBlockOption{}
	} else if _, ok := block.(RestReadManyBlock); ok {
		return &RestBlockOption{}
	} else if _, ok := block.(RestReadBlock); ok {
		return &RestBlockOption{}
	}
	return &CommonBlokOption{}
}
