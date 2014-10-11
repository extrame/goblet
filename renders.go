package goblet

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type RenderInstance interface {
	render(wr http.ResponseWriter, data interface{}) bool
}

type _Render interface {
	render(*BlockOption) RenderInstance
	Init(*Server)
}

type HtmlRender struct {
	root     *template.Template
	dir      string
	models   map[string]*template.Template
	suffix   string
	saveTemp bool
}

func (h *HtmlRender) render(opt *BlockOption) RenderInstance {
	if opt.isHtml {
		layout := h.getTemplate("layout/"+opt.layout+h.suffix, filepath.Join("layout", opt.layout+h.suffix))
		yield := h.getTemplate(opt.htmlRenderFileOrDir + h.suffix)
		return &HttpRenderInstance{layout, yield}
	}
	return nil
}

func (h *HtmlRender) Init(s *Server) {
	h.root = template.New("REST_HTTP_ROOT")
	h.root.Funcs(template.FuncMap{"raw": RawHtml, "yield": RawHtml})
	h.dir = *s.WwwRoot
	h.initGlobalTemplate(h.dir)
	h.initLayout(h.dir)
	h.models = make(map[string]*template.Template)
	h.suffix = ".html"
	h.saveTemp = (*s.env == "production")
}

func (f *HtmlRender) initGlobalTemplate(dir string) {
	f.root.New("")
	//scan for the helpers
	filepath.Walk(filepath.Join(dir, "helper"), func(path string, info os.FileInfo, err error) error {
		if err == nil && (!info.IsDir()) && strings.HasSuffix(info.Name(), "html") {
			fmt.Println("Parse helper:", path)
			name := strings.TrimSuffix(info.Name(), "html")
			e := parseFileWithName(f.root, "global/"+name, path)
			if e != nil {
				fmt.Printf("ERROR template.ParseFile: %v", e)
			}
		}
		return nil
	})
}

func (f *HtmlRender) initLayout(dir string) {
	f.root.New("")
	//scan for the layout
	filepath.Walk(filepath.Join(dir, "helper"), func(path string, info os.FileInfo, err error) error {
		if err == nil && (!info.IsDir()) && strings.HasSuffix(info.Name(), "html") {
			fmt.Println("Parse helper:", path)
			name := strings.TrimSuffix(info.Name(), "html")
			e := parseFileWithName(f.root, "layout/"+name, path)
			if e != nil {
				fmt.Printf("ERROR template.ParseFile: %v", e)
			}
		}
		return nil
	})
}

func (h *HtmlRender) getTemplate(args ...string) *template.Template {
	var name, file string
	if len(args) == 1 {
		name = args[0]
		file = args[0]
	} else {
		name = args[1]
		file = args[1]
	}
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
				fmt.Println("ERROR template.ParseFile: %v", err)
			}
		}
	}
	return t
}

type HttpRenderInstance struct {
	layout *template.Template
	yield  *template.Template
}

func (h *HttpRenderInstance) render(wr http.ResponseWriter, data interface{}) bool {
	h.layout.Funcs(template.FuncMap{
		"yield": func() (template.HTML, error) {
			err := h.yield.Execute(wr, data)
			// return safe html here since we are rendering our own template
			return template.HTML(""), err
		},
	})
	h.layout.Execute(wr, data)
	return true
}

type JsonRender struct {
}

func (j *JsonRender) render(*BlockOption) RenderInstance {
	return nil
}

func (j *JsonRender) Init(s *Server) {

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
