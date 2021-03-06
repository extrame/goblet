package goblet

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	ge "github.com/extrame/goblet/error"
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
		return h.BasicBlockOption.callMethodForBlock("Get", c)
	} else if method == "post" {
		return h.BasicBlockOption.callMethodForBlock("Post", c)
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

func (r *RestBlockOption) renderAsRead(id string, ctx *Context) error {
	return r.BasicBlockOption.callMethodForBlock("Read", ctx)
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
	var args []string
	if len(c.suffix) > 0 {
		id = c.suffix[1:]
		args = strings.SplitN(id, "/", 2)
		id = args[0]
	}
	if id != "" {
		if id == "new" {
			c.method = REST_NEW
			if args != nil && len(args) > 1 {
				c.suffix = args[1]
			}
			return r.BasicBlockOption.callMethodForBlock("New", c)
		} else if method == "GET" || method == "HEAD" {
			if nsuff := strings.TrimSuffix(c.suffix, ";edit"); nsuff != c.suffix {
				c.method = REST_EDIT
				c.suffix = nsuff
				return r.BasicBlockOption.callMethodForBlock("Edit", c)
			} else {
				c.method = REST_READ
				return r.BasicBlockOption.callMethodForBlock("Read", c)
			}
		} else if method == "DELETE" {
			c.method = REST_DELETE
			return r.BasicBlockOption.callMethodForBlock("Delete", c)
		} else {
			c.method = REST_UPDATE
			return r.BasicBlockOption.callMethodForBlock("Update", c)
		}
	} else {
		if method == "GET" {
			c.method = REST_READMANY
			return r.BasicBlockOption.callMethodForBlock("ReadMany", c)
		} else if method == "POST" {
			c.method = REST_CREATE
			return r.BasicBlockOption.callMethodForBlock("Create", c)
		} else if method == "PUT" {
			c.method = REST_UPDATEMANY
			return r.BasicBlockOption.callMethodForBlock("UpdateMany", c)
		} else if method == "DELETE" {
			c.method = REST_DELETEMANY
			return r.BasicBlockOption.callMethodForBlock("DeleteMany", c)
		}
	}

	return nil
}

func (r *BasicBlockOption) tryPre(m string, ctx *Context) bool {
	key := r.name + "-" + m
	key = strings.ToLower(key)
	if pc, ok := ctx.Server.pres[key]; ok {
		for _, fn := range pc {
			results, _ := callMethod(fn, ctx)
			if err, ok := results[0].Interface().(error); ok && err != nil {
				if err != Interrupted {
					ctx.RespondError(err)
				}
				return false
			}
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

		args := strings.Split(name, "/")

		typ := g.block.Type()

		if g.ignoreCase {
			for i := 0; i < g.block.NumMethod(); i++ {
				m := typ.Method(i)
				if strings.ToLower(m.Name) == strings.ToLower(args[0]) {
					name = strings.ToLower(args[0])
					method = g.block.Method(i)
				}
			}
		} else {
			method = g.block.MethodByName(args[0])
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
		} else {
			if len(args) > 1 {
				ctx.suffix = strings.Join(args[1:], "/")
			} else {
				ctx.suffix = ""
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

		if g.tryPre(name, ctx) {
			results, typ := callMethod(method, ctx)
			checkResult(results, typ, ctx)
		}

		// key := strings.ToLower(g.name + "-" + name)
		// if pc, ok := ctx.Server.pres[key]; ok {
		// 	results := callMethod(pc, ctx)
		// 	// pc.Call([]reflect.Value{arg})
		// 	if err, ok := results[0].Interface().(error); ok && err != nil {
		// 		if err != Interrupted {
		// 			ctx.RespondError(err)
		// 		}
		// 		return nil
		// 	}
		// } else {
		// 	// method.Call([]reflect.Value{arg})
		// 	callMethod(method, ctx)
		// }
	}
	return nil

}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func checkResult(results []reflect.Value, typ reflect.Type, ctx *Context) {
	// status_code is not setted
	if len(results) > 0 && !ctx.already_writed && ctx.response == nil && ctx.responseMap == nil {
		for i := len(results); i > 0; i-- {
			var ires = results[i-1].Interface()
			var ti = typ.Out(i - 1)
			ok := ti.Implements(errorInterface)
			if !ok {
				ctx.Respond(ires)
				return
			} else if ok && !results[i-1].IsNil() {
				//优先按最后一个error参数返回错误，符合主流编程习惯
				ctx.RespondError(ires.(error))
				return
			}
		}
		ctx.RespondOK()
	}
}

func callMethod(method reflect.Value, ctx *Context) ([]reflect.Value, reflect.Type) {
	typ := method.Type()
	rvArgs := make([]reflect.Value, typ.NumIn())
	var i = 0
	var suffix = ctx.suffix

	if len(suffix) > 0 && suffix[0] == '/' {
		suffix = suffix[1:]
	}

	for ; i < typ.NumIn(); i++ {
		argT := typ.In(i)
		var kind = argT.Kind()
		if kind == reflect.String || (kind >= reflect.Int && kind <= reflect.Int64) {
			args := strings.SplitN(suffix, "/", 2)
			var newV = reflect.New(argT)

			if kind == reflect.String {
				newV.Elem().SetString(args[0])
			} else {
				iValue, _ := strconv.ParseInt(args[0], 10, 64)
				newV.Elem().SetInt(iValue)
			}

			rvArgs[i] = newV.Elem()

			if len(args) >= 2 {
				suffix = args[1]
			} else {
				suffix = ""
			}
		} else if kind == reflect.Slice && argT.Elem().Kind() == reflect.String {
			args := strings.SplitN(suffix, "/", -1)
			rvArgs[i] = reflect.ValueOf(args)
			i++
			break
		} else {
			break
		}
	}

	rvArgs[i] = reflect.ValueOf(ctx)
	i++

	if i < typ.NumIn() {
		var typArg = typ.In(i)
		var newV reflect.Value
		if typArg.Kind() == reflect.Ptr {
			newV = reflect.New(typ.In(i).Elem())
		} else {
			newV = reflect.New(typ.In(i))
		}

		if err := ctx.Fill(newV.Interface()); err != nil {
			glog.Errorln("parse arguments error", err)
		}
		if typArg.Kind() == reflect.Ptr {
			rvArgs[i] = newV
		} else {
			rvArgs[i] = newV.Elem()
		}

	}

	return method.Call(rvArgs), typ
}

func (r *BasicBlockOption) callMethodForBlock(methodName string, ctx *Context) error {
	method := r.block.MethodByName(methodName)
	if !method.IsValid() {
		var err = fmt.Errorf("you have no method named (%s)", methodName)
		if ctx.Server.Env() == ProductEnv {
			glog.Infof(err.Error())
		} else {
			glog.Fatalf(err.Error())
		}
		return err
	} else {
		if r.tryPre(methodName, ctx) {
			results, typ := callMethod(method, ctx)
			//可以接收传统的无返回，直接结束
			// 或者有返回，如果返回不是error，且不为空，返回结果
			// 如果有返回，且返回是error，不为空，返回错误
			// 其他情况，直接返回ok
			checkResult(results, typ, ctx)
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

func (s *Server) prepareOption(block interface{}) BlockOption {

	var basic BasicBlockOption
	basic.block = reflect.ValueOf(block)

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
		basic.render = []string{s.defaultRender}
	}

	if basic.htmlRenderFileOrDir == "" {
		basic.htmlRenderFileOrDir = strings.ToLower(valtype.Name())
	}

	if basic.layout == "" {
		basic.layout = "default"
	}

	glog.Errorf("[%T]fork on %v", block, basic.routing)

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
		case "Init":
			continue
		default:
			return &groupBlockOption{basic, ignoreCase}
		}
	}
	return &groupBlockOption{basic, ignoreCase}
}
