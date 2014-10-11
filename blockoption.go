package goblet

import (
	"reflect"
	"strings"
)

type Route byte
type Render byte
type Layout byte

type BlockOption struct {
	routing             []string
	render              []string
	layout              string
	isHtml              bool
	htmlRenderFileOrDir string
	isRest              bool
}

func PrepareOption(block interface{}) *BlockOption {
	option := new(BlockOption)

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
				option.routing = tags
				continue
			}
			if t.Type.Name() == "Render" && t.Type.PkgPath() == "github.com/extrame/goblet" {
				option.render = make([]string, len(tags))
				for k, v := range tags {
					vs := strings.Split(v, "=")
					option.render[k] = vs[0]
					if vs[0] == "html" && len(vs) >= 2 {
						option.htmlRenderFileOrDir = vs[1]
					}
				}
				continue
			}
			if t.Type.Name() == "Layout" && t.Type.PkgPath() == "github.com/extrame/goblet" {
				option.layout = string(t.Tag)
				continue
			}
		}
	}

	if len(option.routing) == 0 {
		option.routing = []string{"/" + valtype.Name()}
	}

	if len(option.render) == 0 {
		option.routing = []string{"json"}
	}

	if option.htmlRenderFileOrDir == "" {
		option.htmlRenderFileOrDir = valtype.Name()
	}

	if option.layout == "" {
		option.layout = "default"
	}

	if _, ok := block.(HtmlGetBlock); ok {
		option.isHtml = true
	}

	if _, ok := block.(HtmlPostBlock); ok {
		option.isHtml = true
	}

	if _, ok := block.(RestNewBlock); ok {
		option.isRest = true
	}

	return option
}
