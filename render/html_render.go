package render

import (
	"errors"
	"fmt"
	"github.com/extrame/goblet/error"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type HtmlRender struct {
	root     *template.Template
	dir      string
	models   map[string]*template.Template
	suffix   string
	saveTemp bool
}

func (h *HtmlRender) PrepareInstance(ctx RenderContext) (instance RenderInstance, err error) {
	var layout, yield *template.Template

	err = errors.New("")

	var root *template.Template

	if !h.saveTemp {
		root, _ = h.root.Clone()
		h.initGlobalTemplate(root, h.dir)
	} else {
		root = h.root
	}

	if ctx.StatusCode() >= 300 {
		layout, err = h.getTemplate(root, "layout/"+"error"+h.suffix, filepath.Join("layout", "error"+h.suffix))
		if err != nil {
			layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix))
		}
		yield, err = h.getTemplate(root, strconv.Itoa(ctx.StatusCode())+h.suffix, filepath.Join(strconv.Itoa(ctx.StatusCode())+h.suffix))
		if err != nil {
			log.Println("Find Err Code Fail, ", err)
		}
	}
	if err != nil {
		switch ctx.BlockOptionType() {

		case "Html":
			layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix))
			if err == nil {
				yield, err = h.getTemplate(root, ctx.Method()+h.suffix)
			}
		case "Rest":
			if layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+h.suffix)); err != nil {
				layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix))
			}
			if err == nil {
				h.initModelTemplate(layout, ctx.TemplatePath())
				yield, err = h.getTemplate(root, ctx.TemplatePath()+"/"+ctx.Method()+h.suffix)
			}
		case "Group":
			if layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+h.suffix)); err != nil {
				layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix))
			}
			if err == nil {
				h.initModelTemplate(layout, ctx.TemplatePath())
				yield, err = h.getTemplate(root, ctx.TemplatePath()+"/"+ctx.Method()+h.suffix)
			}
		case "Static":
			if layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+h.suffix)); err != nil {
				layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix))
			}
			if err == nil {
				h.initModelTemplate(layout, ctx.TemplatePath())
				yield, err = h.getTemplate(root, ctx.TemplatePath()+"/"+ctx.Method()+h.suffix)
			}
		}
	}
	if err == nil {
		return &HttpRenderInstance{layout, yield}, nil
	}

	return
}

func (h *HtmlRender) Init(s RenderServer) {
	h.root = template.New("REST_HTTP_ROOT")
	h.root.Funcs(template.FuncMap{"raw": RawHtml, "yield": RawHtml, "status": RawHtml, "slice": Slice, "mask": RawHtml, "repeat": Repeat})
	h.dir = s.WwwRoot()
	h.suffix = ".html"
	h.models = make(map[string]*template.Template)
	h.saveTemp = (s.Env() == "production")
	if h.saveTemp {
		h.initGlobalTemplate(h.root, h.dir)
	}
}

func (h *HtmlRender) initTemplate(parent *template.Template, dir string, typ string) {
	parent.New("")
	if !h.saveTemp { //for debug
		log.Println("init template in ", filepath.Join(h.dir, dir, "helper"))
	}
	//scan for the helpers
	filepath.Walk(filepath.Join(h.dir, dir, "helper"), func(path string, info os.FileInfo, err error) error {
		if err == nil && (!info.IsDir()) && strings.HasSuffix(info.Name(), h.suffix) {
			name := strings.TrimSuffix(info.Name(), h.suffix)
			log.Printf("Parse helper:%s(%s)", typ+"/"+name, path)
			e := parseFileWithName(parent, typ+"/"+name, path)
			if e != nil {
				fmt.Printf("ERROR template.ParseFile: %v", e)
			}
		}
		return nil
	})
}

func (h *HtmlRender) initGlobalTemplate(parent *template.Template, dir string) {
	h.initTemplate(parent, dir, "global")
}

func (h *HtmlRender) initModelTemplate(parent *template.Template, dir string) {
	h.initTemplate(parent, dir, "model")
}

func (h *HtmlRender) getTemplate(root *template.Template, args ...string) (*template.Template, error) {
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
		cloned_rest_model, err := root.Clone()

		if err == nil {

			err = parseFileWithName(cloned_rest_model, name, filepath.Join(h.dir, file))
			if err == nil {
				t = cloned_rest_model.Lookup(name)
				if h.saveTemp {
					h.models[name] = t
				}
			} else {
				if os.IsNotExist(err) {
					return nil, ge.NOSUCHROUTER
				} else {
					return nil, err
				}
			}
		}
	}
	return t, nil
}

type HttpRenderInstance struct {
	layout *template.Template
	yield  *template.Template
}

func (h *HttpRenderInstance) Render(wr http.ResponseWriter, data interface{}, status int) error {
	var mask_map = make(map[string]bool)

	funcMap := template.FuncMap{
		"yield": func() (template.HTML, error) {
			err := h.yield.Execute(wr, data)
			// return safe html here since we are rendering our own template
			return template.HTML(""), err
		},
		"status": func() int {
			return status
		},
		"mask": func(tag string) string {
			if _, ok := mask_map[tag]; ok {
				return "true"
			} else {
				mask_map[tag] = true
			}
			return ""
		},
	}
	h.layout.Funcs(funcMap)
	h.yield.Funcs(funcMap)

	if h.layout != nil {
		return h.layout.Execute(wr, data)
	} else if h.yield != nil {
		return h.yield.Execute(wr, data)
	}
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

func Slice(obj interface{}, leng int) interface{} {
	slice := reflect.ValueOf(obj)
	new_leng := slice.Len() / leng

	if slice.Len()%leng != 0 {
		new_leng++
	}
	new_array := reflect.MakeSlice(reflect.SliceOf(slice.Type()), new_leng, new_leng)
	for i := 0; i < new_leng; i++ {
		end := (i + 1) * leng
		if end > slice.Len() {
			end = slice.Len()
		}
		item_array_in_new_array := slice.Slice(i*leng, end)
		new_array.Index(i).Set(item_array_in_new_array)
	}
	return new_array.Interface()
}

func Repeat(count int) []int {
	res := make([]int, count)
	for i := 0; i < count; i++ {
		res[i] = i
	}
	return res
}
