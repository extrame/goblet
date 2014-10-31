package goblet

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type RenderInstance interface {
	render(wr http.ResponseWriter, data interface{}, status int) error
}

type _Render interface {
	render(*Context) RenderInstance
	Init(*Server)
}

type HtmlRender struct {
	root     *template.Template
	dir      string
	models   map[string]*template.Template
	suffix   string
	saveTemp bool
}

func (h *HtmlRender) render(ctx *Context) RenderInstance {
	var err = errors.New("")
	var layout, yield *template.Template

	if ctx.status_code != 200 {
		layout, err = h.getTemplate("layout/"+"error"+h.suffix, filepath.Join("layout", "error"+h.suffix))
		if err == nil {
			yield, err = h.getTemplate(strconv.Itoa(ctx.status_code)+h.suffix, filepath.Join(strconv.Itoa(ctx.status_code)+h.suffix))
		}
	}
	if err != nil {
		switch typ := ctx.option.(type) {

		case *HtmlBlockOption:
			layout, err = h.getTemplate("layout/"+ctx.getLayout()+h.suffix, filepath.Join("layout", ctx.getLayout()+h.suffix))
			if err == nil {
				yield, err = h.getTemplate(typ.htmlRenderFileOrDir + h.suffix)
			}
		case *RestBlockOption:
			if layout, err = h.getTemplate("layout/"+ctx.getLayout()+h.suffix, filepath.Join(typ.htmlRenderFileOrDir, "layout", ctx.getLayout()+h.suffix)); err != nil {
				layout, err = h.getTemplate("layout/"+ctx.getLayout()+h.suffix, filepath.Join("layout", ctx.getLayout()+h.suffix))
			}
			h.initModelTemplate(layout, typ.htmlRenderFileOrDir)
			if err == nil {
				yield, err = h.getTemplate(typ.htmlRenderFileOrDir + "/" + ctx.method + h.suffix)
			}
		}
	}

	if err == nil {
		return &HttpRenderInstance{layout, yield}
	} else {
		fmt.Println(err)
	}

	return nil
}

func (h *HtmlRender) Init(s *Server) {
	h.root = template.New("REST_HTTP_ROOT")
	h.root.Funcs(template.FuncMap{"raw": RawHtml, "yield": RawHtml})
	h.dir = *s.WwwRoot
	h.suffix = ".html"
	h.initGlobalTemplate(h.dir)
	h.models = make(map[string]*template.Template)
	h.saveTemp = (*s.env == "production")
}

func (f *HtmlRender) initGlobalTemplate(dir string) {
	f.root.New("")
	//scan for the helpers
	filepath.Walk(filepath.Join(f.dir, dir, "helper"), func(path string, info os.FileInfo, err error) error {
		if err == nil && (!info.IsDir()) && strings.HasSuffix(info.Name(), f.suffix) {
			fmt.Println("Parse helper:", path)
			name := strings.TrimSuffix(info.Name(), f.suffix)
			e := parseFileWithName(f.root, "global/"+name, path)
			if e != nil {
				fmt.Printf("ERROR template.ParseFile: %v", e)
			}
		}
		return nil
	})
}

func (h *HtmlRender) initModelTemplate(layout *template.Template, dir string) {
	layout.New("")
	//scan for the helpers
	filepath.Walk(filepath.Join(h.dir, dir, "helper"), func(path string, info os.FileInfo, err error) error {
		if err == nil && (!info.IsDir()) && strings.HasSuffix(info.Name(), h.suffix) {
			fmt.Println("Parse helper:", path)
			name := strings.TrimSuffix(info.Name(), h.suffix)
			e := parseFileWithName(layout, "model/"+name, path)
			if e != nil {
				fmt.Printf("ERROR template.ParseFile: %v", e)
			}
		}
		return nil
	})
}

func (h *HtmlRender) getTemplate(args ...string) (*template.Template, error) {
	var name, file string
	if len(args) == 1 {
		name = args[0]
		file = args[0]
	} else {
		name = args[1]
		file = args[1]
	}
	file = filepath.FromSlash(file)
	t := h.models[name]

	if t == nil {
		cloned_rest_model, err := h.root.Clone()

		if err == nil {

			err = parseFileWithName(cloned_rest_model, name, filepath.Join(h.dir, file))
			if err == nil {
				t = cloned_rest_model.Lookup(name)
				if h.saveTemp {
					h.models[name] = t
				}
			} else {
				return nil, err
			}
		}
	}
	return t, nil
}

type HttpRenderInstance struct {
	layout *template.Template
	yield  *template.Template
}

func (h *HttpRenderInstance) render(wr http.ResponseWriter, data interface{}, status int) error {
	if h.layout != nil {
		h.layout.Funcs(template.FuncMap{
			"yield": func() (template.HTML, error) {
				err := h.yield.Execute(wr, data)
				// return safe html here since we are rendering our own template
				return template.HTML(""), err
			},
		})
		return h.layout.Execute(wr, data)
	} else if h.yield != nil {
		return h.yield.Execute(wr, data)
	}
	return nil
}

type JsonRender struct {
}

func (j *JsonRender) render(c *Context) RenderInstance {
	return new(JsonRenderInstance)
}

func (j *JsonRender) Init(s *Server) {

}

type JsonRenderInstance int8

func (r *JsonRenderInstance) render(wr http.ResponseWriter, data interface{}, status int) (err error) {
	var v []byte
	v, err = json.Marshal(data)
	if err == nil {
		wr.Write(v)
	}
	wr.WriteHeader(status)
	return
}

type RawRender int8

func (r *RawRender) render(c *Context) RenderInstance {
	return new(RawRenderInstance)
}

func (r *RawRender) Init(s *Server) {
}

type RawRenderInstance int8

func (r *RawRenderInstance) render(wr http.ResponseWriter, data interface{}, status int) error {
	return nil
}

func parseFileWithName(parent *template.Template, name string, filepath string) error {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	s := string(b)
	// First template becomes return value if not already defined,
	// and we use that one for subsequent New calls to associate
	// all the templates together. Also, if this file has the same name
	// as t, this file becomes the contents of t, so
	//  t, err := New(name).Funcs(xxx).ParseFiles(name)
	// works. Otherwise we create a new template associated with t.
	var tmpl *template.Template
	if name == parent.Name() || name == "" {
		tmpl = parent
	} else {
		tmpl = parent.New(name)
	}
	_, err = tmpl.Parse(s)
	if err != nil {
		return err
	}
	return nil
}

func RawHtml(text string) template.HTML { return template.HTML(text) }
