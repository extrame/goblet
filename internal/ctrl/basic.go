package ctrl

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/extrame/goblet/config"
	"github.com/extrame/goblet/internal/errors"
	"github.com/extrame/goblet/internal/matcher"
)

type Server interface {
	GetDefaultRender() string
}

type MethodCaller struct {
	fn reflect.Value
}

func (m *MethodCaller) MatchSuffix(string) bool {
	return true
}

func (m *MethodCaller) String() string {
	return fmt.Sprintf("%v", m.fn)
}

func (m *MethodCaller) IsValid() bool {
	return m.fn.IsValid()
}

func DetectOption(ctrl interface{}, server Server) (basic *Basic, ignoreCase bool) {
	basic = &Basic{}
	basic.block = reflect.ValueOf(ctrl)
	ignoreCase = true
	var val reflect.Value
	var valtypeOrigin, valtype reflect.Type
	valtypeOrigin = basic.block.Type()

	if basic.block.Kind() == reflect.Ptr {
		val = basic.block.Elem()
	} else {
		val = basic.block
	}
	valtype = val.Type()

	basic.Name = valtype.Name()

	if val.Kind() == reflect.Struct {
		for i := 0; i < valtype.NumField(); i++ {
			t := valtype.Field(i)

			if t.Type.PkgPath() != "github.com/extrame/goblet" {
				continue
			}

			if t.Type.Name() == "Layout" {
				basic.layout = string(t.Tag)
				continue
			}
			if t.Type.Name() == "SingleController" {
				basic.typ = "single"
				continue
			}

			if t.Type.Name() == "RestController" {
				basic.typ = "rest"
				continue
			}

			if t.Type.Name() == "AutoHide" {
				basic.hide = true
				continue
			}
			if t.Type.Name() == "ErrorRender" {
				basic.errRender = string(t.Tag)
				continue
			}

			tags := strings.Split(string(t.Tag), ",")

			if t.Type.Name() == "GroupController" {
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

			if t.Type.Name() == "Route" {
				var tag = t.Tag.Get("path")
				basic.routing = []string{tag}
				basic.htmlRenderFileOrDir = strings.TrimLeft(tag, "/")
				continue
			}

			if t.Type.Name() == "Render" {
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
		// parse methods and store in map
		basic.methods = matcher.Root(nil)
		for i := 0; i < valtypeOrigin.NumMethod(); i++ {
			mtd := valtypeOrigin.Method(i)
			hyphenName := toHyphenCase(mtd.Name)
			hyphenName = strings.ReplaceAll(hyphenName, "_", "/")

			if ignoreCase {
				hyphenName = strings.ToLower(hyphenName)
			}
			basic.methods.Add("/"+hyphenName, &MethodCaller{
				fn: basic.block.Method(i),
			})
		}
	}

	if len(basic.routing) == 0 {
		basic.routing = []string{"/" + strings.ToLower(valtype.Name())}
	}

	if len(basic.render) == 0 {
		basic.render = []string{server.GetDefaultRender()}
	}

	if basic.htmlRenderFileOrDir == "" {
		basic.htmlRenderFileOrDir = strings.ToLower(valtype.Name())
	}

	if basic.layout == "" {
		basic.layout = "default"
	}
	return
}

// Basic 探测的用户控制器的信息
type Basic struct {
	routing             []string
	render              []string
	layout              string
	htmlRenderFileOrDir string
	typ                 string
	block               reflect.Value
	Name                string
	errRender           string
	hide                bool
	methods             *matcher.UrlMatcher
}

func (b *Basic) Layout() string {
	return b.layout
}

func (b *Basic) TemplatePath() string {
	return b.htmlRenderFileOrDir
}

func (h *Basic) UpdateRender(o string, ctx Context) {
	h.htmlRenderFileOrDir = o
}

func (h *Basic) SetRender(renders []string) {
	h.render = renders
}

func (h *Basic) AutoHidden() bool {
	return h.hide
}

func (h *Basic) GetRender() []string {
	return h.render
}

func (b *Basic) GetRouting() []string {
	return b.routing
}

func (b *Basic) ErrorRender() string {
	return b.errRender
}

func (r *Basic) tryPre(m string, ctx Context) bool {
	key := r.Name + "-" + m
	key = strings.ToLower(key)
	if pc := ctx.GetPre(key); pc != nil {
		for _, fn := range pc {
			results, _ := callMethod(fn, ctx)
			var result = results[0].Interface()
			if result != nil {
				switch tr := result.(type) {
				case *errors.InternalInterruptedError:
					ctx.RespondError(tr)
				}
				return false
			}
		}
	}
	return true
}

func (h *Basic) String() string {
	return fmt.Sprintf("Basic(%s)", h.Name)
}

// toHyphenCase 将驼峰命名转换为连字符格式，例如：AaaBbb -> aaa-bbb
func toHyphenCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		result = append(result, r)
	}
	return string(result)
}

func (r *Basic) callMethodForBlock(methodName string, ctx Context) error {
	method := r.block.MethodByName(methodName)
	if !method.IsValid() {
		var err = fmt.Errorf("you have no method named (%s)", methodName)
		if ctx.Env() == config.ProductEnv {
			slog.Info("block option error", "error", err)
		} else {
			slog.Error("block option fatal error", "error", err)
			os.Exit(1)
		}

		return err
	} else {
		if r.tryPre(methodName, ctx) {
			results, typ := callMethod(method, ctx)
			//可以接收传统的无返回，直接结束
			// 或者有返回，如果返回不是error，且不为空，返回结果
			// 如果有返回，且返回是error，不为空，返回错误
			// 其他情况，直接返回ok
			return checkResult(results, typ, ctx)
		}
	}

	return nil
}
