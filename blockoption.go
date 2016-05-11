package goblet

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/extrame/goblet/error"
)

type Route byte
type Render byte
type Layout byte
type SingleController byte
type RestController byte
type GroupController byte

const (
	REST_READ       = "read"
	REST_READMANY   = "readmany"
	REST_DELETE     = "delete"
	REST_EDIT       = "edit"
	REST_NEW        = "new"
	REST_CREATE     = "create"
	REST_UPDATE     = "update"
	REST_UPDATEMANY = "updatemany"
	REST_DELETEMANY = "deletemany"
)

var RenderNotAllowd = fmt.Errorf("render is not allowed")

type BlockOption interface {
	UpdateRender(string, *Context)
	GetRouting() []string
	MatchSuffix(string) bool
	//Get the render by user require, if required render is not allow, pass RenderNotAllowed
	GetRender(*Context) (render string, err error)
	//Reset the allowed renders
	SetRender([]string)
	//Call the function in object and Parse data, this function used before
	//the render prepared. So you can change function and render in here
	Parse(*Context) error
	Layout() string
	TemplatePath() string
}

type BasicBlockOption struct {
	routing             []string
	render              []string
	layout              string
	htmlRenderFileOrDir string
	typ                 string
	block               interface{}
	name                string
}

type HtmlBlockOption struct {
	BasicBlockOption
}

func (h *HtmlBlockOption) MatchSuffix(suffix string) bool {
	return len(suffix) == 0
}

func (h *HtmlBlockOption) Parse(c *Context) error {
	c.method = h.htmlRenderFileOrDir
	if c.lCtx.ReqMethod() == "GET" {
		if get, ok := h.BasicBlockOption.block.(HtmlGetBlock); ok {
			get.Get(c)
		}
	} else if c.lCtx.ReqMethod() == "POST" {
		if post, ok := h.BasicBlockOption.block.(HtmlPostBlock); ok {
			post.Post(c)
		}
	}
	return nil
}

func (b *BasicBlockOption) Layout() string {
	return b.layout
}

func (b *BasicBlockOption) TemplatePath() string {
	return b.htmlRenderFileOrDir
}

func (h *BasicBlockOption) UpdateRender(o string, ctx *Context) {
	h.htmlRenderFileOrDir = o
}

func (h *BasicBlockOption) SetRender(renders []string) {
	h.render = renders
}

func (b *BasicBlockOption) GetRouting() []string {
	return b.routing
}

func (b *BasicBlockOption) GetRender(cx *Context) (render string, err error) {
	if cx.forceFormat != "" {
		return cx.forceFormat, nil
	}
	if cx.format == "" {
		return b.render[0], nil
	} else {
		for _, v := range b.render {
			if v == cx.format {
				return v, nil
			}
		}
		if cx.tempRenders != nil {
			for _, v := range cx.tempRenders {
				if v == cx.format {
					return v, nil
				}
			}
		}
	}
	return "", RenderNotAllowd
}

type RestBlockOption struct {
	BasicBlockOption
}

func (r *RestBlockOption) UpdateRender(obj string, ctx *Context) {
	ctx.method = obj
}

func (r *RestBlockOption) Parse(c *Context) error {
	var method string
	if method = c.lCtx.QueryValue("method"); method == "" {
		method = c.lCtx.ReqMethod()
	}
	method = strings.ToUpper(method)
	if len(c.suffix) > 0 {
		id := c.suffix[1:]
		if id == "new" {
			r.renderAsNew(c)
		} else if method == "GET" {
			if nid := strings.TrimSuffix(id, ";edit"); nid != id {
				r.renderAsEdit(nid, c)
			} else {
				r.renderAsRead(id, c)
			}
		} else if method == "DELETE" {
			r.renderAsDelete(id, c)
		} else {
			r.renderAsUpdate(id, c)
		}
	} else {
		if method == "GET" {
			r.renderAsReadMany(c)
		} else if method == "POST" {
			r.renderAsCreate(c)
		} else if method == "PUT" {
			r.renderAsUpdateMany(c)
		} else if method == "DELETE" {
			r.renderAsDeleteMany(c)
		}
	}

	return nil
}

func (r *RestBlockOption) tryPre(m string, ctx *Context) bool {
	arg := reflect.ValueOf(ctx)
	key := r.name + "-" + m
	key = strings.ToLower(key)
	if pc, ok := ctx.Server.pres[key]; ok {
		results := pc.Call([]reflect.Value{arg})
		if err, ok := results[0].Interface().(error); ok && err != nil {
			ctx.RespondError(err)
			return false
		}
	}
	return true
}

func (r *RestBlockOption) renderAsRead(id string, cx *Context) {
	if reader, ok := r.BasicBlockOption.block.(RestReadBlock); ok {
		cx.method = REST_READ
		if r.tryPre("Read", cx) {
			reader.Read(id, cx)
		}
	}
}

func (r *RestBlockOption) renderAsReadMany(cx *Context) {
	if reader, ok := r.BasicBlockOption.block.(RestReadManyBlock); ok {
		cx.method = REST_READMANY
		if r.tryPre("ReadMany", cx) {
			reader.ReadMany(cx)
		}
	}
}

func (r *RestBlockOption) renderAsNew(cx *Context) {
	if reader, ok := r.BasicBlockOption.block.(RestNewBlock); ok {
		cx.method = REST_NEW
		if r.tryPre("New", cx) {
			reader.New(cx)
		}
	}
}

func (r *RestBlockOption) renderAsUpdateMany(cx *Context) {
	if um, ok := r.BasicBlockOption.block.(RestUpdateManyBlock); ok {
		cx.method = REST_UPDATEMANY
		if r.tryPre("UpdateMany", cx) {
			um.UpdateMany(cx)
		}
	}
}

func (r *RestBlockOption) renderAsDeleteMany(cx *Context) {
	if um, ok := r.BasicBlockOption.block.(RestDeleteManyBlock); ok {
		cx.method = REST_DELETEMANY
		if r.tryPre("DeleteMany", cx) {
			um.DeleteMany(cx)
		}
	}
}

func (r *RestBlockOption) renderAsCreate(cx *Context) {
	if um, ok := r.BasicBlockOption.block.(RestCreateBlock); ok {
		cx.method = REST_CREATE
		if r.tryPre("Create", cx) {
			um.Create(cx)
		}
	}
}

func (r *RestBlockOption) renderAsEdit(id string, cx *Context) {
	if um, ok := r.BasicBlockOption.block.(RestEditBlock); ok {
		cx.method = REST_EDIT
		if r.tryPre("Edit", cx) {
			um.Edit(id, cx)
		}
	}
}

func (r *RestBlockOption) renderAsUpdate(id string, cx *Context) {
	if um, ok := r.BasicBlockOption.block.(RestUpdateBlock); ok {
		cx.method = REST_UPDATE
		if r.tryPre("Update", cx) {
			um.Update(id, cx)
		}
	}
}

func (r *RestBlockOption) renderAsDelete(id string, cx *Context) {
	if um, ok := r.BasicBlockOption.block.(RestDeleteBlock); ok {
		cx.method = REST_DELETE
		if r.tryPre("Delete", cx) {
			um.Delete(id, cx)
		}
	}
}

func (r *RestBlockOption) handleData(c *Context) {

}

func (r *RestBlockOption) MatchSuffix(suffix string) bool {
	return len(suffix) == 0 || suffix[0:1] == "/"
}

type GroupBlockOption struct {
	BasicBlockOption
	ignoreCase bool
}

func (c *GroupBlockOption) MatchSuffix(suffix string) bool {
	return true
}

func (g *GroupBlockOption) Parse(ctx *Context) error {
	val := reflect.ValueOf(g.block)
	var method reflect.Value
	var name string
	if len(ctx.suffix) > 1 {
		name = ctx.suffix[1:]

		typ := val.Type()

		if g.ignoreCase {
			for i := 0; i < val.NumMethod(); i++ {
				m := typ.Method(i)
				if strings.ToLower(m.Name) == strings.ToLower(name) {
					name = strings.ToLower(name)
					method = val.Method(i)
				}
			}
		} else {
			method = val.MethodByName(name)
		}
		if !method.IsValid() {
			if name = ctx.lCtx.QueryValue("method"); name == "" {
				name = ctx.lCtx.ReqMethod()
			}
			name = strings.ToLower(name)
			switch name {
			case "post":
				method = val.MethodByName("Post")
			case "get":
				method = val.MethodByName("Get")
			}
		}
	}

	if !method.IsValid() {
		if name = ctx.lCtx.QueryValue("method"); name == "" {
			name = ctx.lCtx.ReqMethod()
		}
		name = strings.ToLower(name)
		switch name {
		case "post":
			method = val.MethodByName("Post")
		case "get":
			method = val.MethodByName("Get")
		}
	}
	if !method.IsValid() {
		return ge.NOSUCHROUTER
	} else {
		ctx.method = name
		arg := reflect.ValueOf(ctx)
		key := strings.ToLower(g.name + "-" + name)
		if pc, ok := ctx.Server.pres[key]; ok {
			results := pc.Call([]reflect.Value{arg})
			if err, ok := results[0].Interface().(error); ok && err != nil {
				ctx.RespondError(err)
				return nil
			}
		} else {
			method.Call([]reflect.Value{arg})
		}
	}
	return nil

}

type _staticBlockOption struct {
	BasicBlockOption
}

func (c *_staticBlockOption) MatchSuffix(suffix string) bool {
	return true
}

func (c *_staticBlockOption) Parse(ctx *Context) error {
	if len(ctx.suffix) > 1 {
		ctx.method = ctx.suffix
	} else {
		ctx.method = "index"
	}
	ctx.forceFormat = "html"
	ctx.layout = "default"
	return nil
}

func PrepareOption(block interface{}) BlockOption {

	var basic BasicBlockOption
	basic.block = block
	val := reflect.ValueOf(block)

	// initMethod := val.MethodByName("Init")
	// if initMethod.IsValid() {
	// 	initMethod.Call(nil)
	// }

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	basic.name = val.Type().Name()
	valtype := val.Type()
	ignoreCase := true

	if val.Kind() == reflect.Struct {
		for i := 0; i < valtype.NumField(); i++ {
			t := valtype.Field(i)

			if t.Type.Name() == "Layout" && t.Type.PkgPath() == "github.com/extrame/goblet" {
				basic.layout = string(t.Tag)
				continue
			}
			if t.Type.Name() == "SingleController" && t.Type.PkgPath() == "github.com/extrame/goblet" {
				basic.typ = "single"
				continue
			}

			if t.Type.Name() == "RestController" && t.Type.PkgPath() == "github.com/extrame/goblet" {
				basic.typ = "rest"
				continue
			}

			tags := strings.Split(string(t.Tag), ",")

			if t.Type.Name() == "GroupController" && t.Type.PkgPath() == "github.com/extrame/goblet" {
				basic.typ = "group"
				for _, v := range tags {
					vs := strings.Split(v, "=")
					if vs[0] == "ignoreCase" && len(vs) >= 2 {
						if vs[1] == "false" {
							ignoreCase = false
						}
					}
				}
				continue
			}

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
		}
	}

	if len(basic.routing) == 0 {
		basic.routing = []string{"/" + strings.ToLower(valtype.Name())}
	}

	if len(basic.render) == 0 {
		basic.render = []string{"html"}
	}

	if basic.htmlRenderFileOrDir == "" {
		basic.htmlRenderFileOrDir = strings.ToLower(valtype.Name())
	}

	if basic.layout == "" {
		basic.layout = "default"
	}

	return newBlock(basic, block, ignoreCase)
}

func newBlock(basic BasicBlockOption, block interface{}, ignoreCase bool) BlockOption {
	switch basic.typ {
	case "single":
		return &HtmlBlockOption{basic}
	case "rest":
		return &RestBlockOption{basic}
	case "group":
		return &GroupBlockOption{basic, ignoreCase}
	}

	switch block.(type) {
	case HtmlGetBlock, HtmlPostBlock:
		return &HtmlBlockOption{basic}
	case RestNewBlock, RestReadManyBlock, RestReadBlock, RestUpdateBlock, RestUpdateManyBlock, RestDeleteBlock, RestDeleteManyBlock, RestEditBlock, RestCreateBlock:
		return &RestBlockOption{basic}
	default:
		return &GroupBlockOption{basic, ignoreCase}
	}

}
