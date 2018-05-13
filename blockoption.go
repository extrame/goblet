package goblet

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/extrame/goblet/error"
	"github.com/golang/glog"
)

type Route byte
type Render byte
type Layout byte
type SingleController byte
type RestController byte
type GroupController byte
type ErrorRender byte
type NoHidden byte

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
	GetRender() (render []string)
	//Reset the allowed renders
	SetRender([]string)
	//Call the function in object and Parse data, this function used before
	//the render prepared. So you can change function and render in here
	Parse(*Context) error
	Layout() string
	TemplatePath() string
	ErrorRender() string
	AutoHidden() bool
}

type BasicBlockOption struct {
	routing             []string
	render              []string
	layout              string
	htmlRenderFileOrDir string
	typ                 string
	block               reflect.Value
	name                string
	errRender           string
	noHidden            bool
}

type HtmlBlockOption struct {
	BasicBlockOption
}

func (h *HtmlBlockOption) MatchSuffix(suffix string) bool {
	return len(suffix) == 0 || len(suffix) == 1 && suffix[0:1] == "/"
}

func (h *HtmlBlockOption) Parse(c *Context) error {
	c.method = h.htmlRenderFileOrDir
	var method string
	if method = c.request.URL.Query().Get("method"); method == "" {
		method = c.request.Method
	}
	method = strings.ToLower(method)

	if method == "get" {
		h.BasicBlockOption.callMethodForBlock("Get", c)
	} else if method == "post" {
		h.BasicBlockOption.callMethodForBlock("Post", c)
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

func (h *BasicBlockOption) AutoHidden() bool {
	return !h.noHidden
}

func (h *BasicBlockOption) GetRender() []string {
	return h.render
}

func (b *BasicBlockOption) GetRouting() []string {
	return b.routing
}

func (b *BasicBlockOption) ErrorRender() string {
	return b.errRender
}

type RestBlockOption struct {
	BasicBlockOption
}

func (r *RestBlockOption) renderAsRead(id string, ctx *Context) {
	r.BasicBlockOption.callMethodForBlock("Read", ctx)
}

func (r *RestBlockOption) UpdateRender(obj string, ctx *Context) {
	ctx.method = obj
}

func (r *RestBlockOption) Parse(c *Context) error {
	var method string
	if method = c.request.URL.Query().Get("method"); method == "" {
		method = c.request.Method
	}
	method = strings.ToUpper(method)
	var id string
	if len(c.suffix) > 0 {
		id = c.suffix[1:]
	}
	if id != "" {
		if id == "new" {
			c.method = REST_NEW
			r.BasicBlockOption.callMethodForBlock("New", c)
		} else if method == "GET" {
			if nsuff := strings.TrimSuffix(c.suffix, ";edit"); nsuff != c.suffix {
				c.method = REST_EDIT
				c.suffix = nsuff
				r.BasicBlockOption.callMethodForBlock("Edit", c)
			} else {
				c.method = REST_READ
				r.BasicBlockOption.callMethodForBlock("Read", c)
			}
		} else if method == "DELETE" {
			c.method = REST_DELETE
			r.BasicBlockOption.callMethodForBlock("Delete", c)
		} else {
			c.method = REST_UPDATE
			r.BasicBlockOption.callMethodForBlock("Update", c)
		}
	} else {
		if method == "GET" {
			c.method = REST_READMANY
			r.BasicBlockOption.callMethodForBlock("ReadMany", c)
		} else if method == "POST" {
			c.method = REST_CREATE
			r.BasicBlockOption.callMethodForBlock("Create", c)
		} else if method == "PUT" {
			c.method = REST_UPDATEMANY
			r.BasicBlockOption.callMethodForBlock("UpdateMany", c)
		} else if method == "DELETE" {
			c.method = REST_DELETEMANY
			r.BasicBlockOption.callMethodForBlock("DeleteMany", c)
		}
	}

	return nil
}

func (r *BasicBlockOption) tryPre(m string, ctx *Context) bool {
	key := r.name + "-" + m
	key = strings.ToLower(key)
	if pc, ok := ctx.Server.pres[key]; ok {
		results := callMethod(pc, ctx)
		if err, ok := results[0].Interface().(error); ok && err != nil {
			if err != Interrupted {
				ctx.RespondError(err)
			}
			return false
		}
	}
	return true
}

func (r *RestBlockOption) handleData(c *Context) {

}

func (r *RestBlockOption) MatchSuffix(suffix string) bool {
	return len(suffix) == 0 || suffix[0:1] == "/"
}

type groupBlockOption struct {
	BasicBlockOption
	ignoreCase bool
}

func (c *groupBlockOption) MatchSuffix(suffix string) bool {
	return true
}

func (g *groupBlockOption) String() string {
	var with = "with"
	if !g.ignoreCase {
		with = "without"
	}
	return fmt.Sprintf("GroupBlockOption %s ignore case", with)
}

func (g *groupBlockOption) Parse(ctx *Context) error {
	var method reflect.Value
	var name string

	if len(ctx.suffix) > 1 {
		name = ctx.suffix[1:]

		typ := g.block.Type()

		if g.ignoreCase {
			for i := 0; i < g.block.NumMethod(); i++ {
				m := typ.Method(i)
				if strings.ToLower(m.Name) == strings.ToLower(name) {
					name = strings.ToLower(name)
					method = g.block.Method(i)
				}
			}
		} else {
			method = g.block.MethodByName(name)
		}
		if !method.IsValid() {
			if name = ctx.request.URL.Query().Get("method"); name == "" {
				name = ctx.request.Method
			}
			name = strings.ToLower(name)
			switch name {
			case "post":
				method = g.block.MethodByName("Post")
			case "get":
				method = g.block.MethodByName("Get")
			}
		}
	}
	if !method.IsValid() {
		if name = ctx.request.URL.Query().Get("method"); name == "" {
			name = ctx.request.Method
		}
		name = strings.ToLower(name)
		switch name {
		case "post":
			method = g.block.MethodByName("Post")
		case "get":
			method = g.block.MethodByName("Get")
		}
	}
	if !method.IsValid() {
		return ge.NOSUCHROUTER
	} else {
		ctx.method = name

		key := strings.ToLower(g.name + "-" + name)
		if pc, ok := ctx.Server.pres[key]; ok {
			results := callMethod(pc, ctx)
			// pc.Call([]reflect.Value{arg})
			if err, ok := results[0].Interface().(error); ok && err != nil {
				if err != Interrupted {
					ctx.RespondError(err)
				}
				return nil
			}
		} else {
			// method.Call([]reflect.Value{arg})
			callMethod(method, ctx)
		}
	}
	return nil

}

func callMethod(method reflect.Value, ctx *Context) []reflect.Value {
	typ := method.Type()
	rvArgs := make([]reflect.Value, typ.NumIn())
	var n, i = 0, 0
	var args []string

	for ; i < typ.NumIn(); i++ {
		argT := typ.In(i)
		if argT.Kind() == reflect.String {
			if args == nil {
				if ctx.suffix[0] == '/' {
					args = strings.Split(ctx.suffix[1:], "/")
				} else {
					args = strings.Split(ctx.suffix, "/")
				}
			}
			if n < len(args) {
				rvArgs[i] = reflect.ValueOf(args[n])
				n++
			} else {
				fmt.Printf("[WARNING]method want more argument(%d) than url part (%v) have\n", typ.NumIn(), args)
			}
		} else {
			break
		}
	}

	rvArgs[i] = reflect.ValueOf(ctx)
	i++

	if i < typ.NumIn() {
		newV := reflect.New(typ.In(i))
		if err := ctx.Fill(newV.Interface()); err != nil {
			glog.Errorln("parse arguments error", err)
		}
		rvArgs[i] = newV.Elem()
	}

	return method.Call(rvArgs)
}

func (r *BasicBlockOption) callMethodForBlock(methodName string, ctx *Context) {
	method := r.block.MethodByName(methodName)
	if !method.IsValid() {
		glog.Fatalf("you have no method named (%s)", methodName)
	}
	if r.tryPre(methodName, ctx) {
		callMethod(method, ctx)
	}
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

func (s *Server) prepareOption(block interface{}) BlockOption {

	var basic BasicBlockOption
	basic.block = reflect.ValueOf(block)
	initMethod := basic.block.MethodByName("Init")
	if initMethod.IsValid() {
		if initMethod.Type().NumIn() == 0 {
			initMethod.Call(nil)
		} else {
			initMethod.Call([]reflect.Value{reflect.ValueOf(s)})
		}
	}

	var val reflect.Value
	var valtype reflect.Type

	if basic.block.Kind() == reflect.Ptr {
		val = basic.block.Elem()
	} else {
		val = basic.block
	}
	valtype = val.Type()

	basic.name = valtype.Name()

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
			if t.Type.Name() == "NoHidden" && t.Type.PkgPath() == "github.com/extrame/goblet" {
				basic.noHidden = true
				continue
			}
			if t.Type.Name() == "ErrorRender" && t.Type.PkgPath() == "github.com/extrame/goblet" {
				basic.errRender = string(t.Tag)
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
		return &groupBlockOption{basic, ignoreCase}
	}

	for i := 0; i < basic.block.Type().NumMethod(); i++ {
		mtd := basic.block.Type().Method(i)
		switch mtd.Name {
		case "Get", "Post":
			return &HtmlBlockOption{basic}
		case "Read", "ReadMany", "Delete", "DeleteMany", "Update", "UpdateMany", "New", "Create", "Edit":
			return &RestBlockOption{basic}
		}
	}
	return &groupBlockOption{basic, ignoreCase}
}
