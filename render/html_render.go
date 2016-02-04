package render

import (
	"fmt"
	"github.com/extrame/goblet/error"
	"github.com/valyala/fasthttp"
	"html/template"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type HtmlRender struct {
	root     *template.Template
	dir      string
	models   map[string]*template.Template
	suffix   string
	public   string
	saveTemp bool
}

func (h *HtmlRender) PrepareInstance(ctx RenderContext) (instance RenderInstance, err error) {
	var layout, yield *template.Template

	var root *template.Template

	var status_code = ctx.StatusCode()

	if !h.saveTemp {
		root, _ = h.root.Clone()
		h.initGlobalTemplate(root)
	} else {
		root = h.root
	}

	path := ctx.TemplatePath() + "/" + ctx.Method()

	h.initModelTemplate(root, ctx.TemplatePath())
	switch ctx.BlockOptionType() {

	case "Html":
		layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix))
		path = ctx.Method()
		if err == nil {
			yield, err = h.getTemplate(root, path+h.suffix)
		}
	case "Rest":
		if layout, err = h.getTemplate(root, "module_layout/"+ctx.Layout()+h.suffix, filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+h.suffix)); err != nil {
			layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix))
		}
		if err == nil {
			yield, err = h.getTemplate(root, path+h.suffix)
		}
	case "Group":
		if layout, err = h.getTemplate(root, "module_layout/"+ctx.Layout()+h.suffix, filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+h.suffix)); err != nil {
			layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix))
		}
		if err == nil {
			yield, err = h.getTemplate(root, path+h.suffix)
		}
	case "Static":
		if layout, err = h.getTemplate(root, "module_layout/"+ctx.Layout()+h.suffix, filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+h.suffix)); err != nil {
			layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix))
		}
		path = ctx.Method()
		if err == nil {
			yield, err = h.getTemplate(root, path+h.suffix)
		}
		//for static file
		if err == ge.NOSUCHROUTER {
			file := filepath.Join(h.dir, h.public, fmt.Sprintf("%s.%s", path, ctx.Format()))
			if _, err := os.Stat(file); os.IsNotExist(err) {
				status_code = 404
			}
		}
	}

	if status_code >= 300 {
		layout, err = h.getTemplate(root, "layout/"+"error"+h.suffix, filepath.Join("layout", "error"+h.suffix))
		if err != nil {
			layout, err = h.getTemplate(root, "layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix))
		}
		yield, err = h.getTemplate(root, strconv.Itoa(status_code)+h.suffix, filepath.Join(strconv.Itoa(status_code)+h.suffix))
		if err != nil {
			log.Println("Find Err Code Fail, ", err)
		}
	}

	if err == nil {
		var charset string
		if charset = ctx.CharSet(); charset == "" {
			charset = "UTF-8"
		}
		return &HttpRenderInstance{layout, yield, "/css/" + path + ".css", "/js/" + path + ".js", charset}, nil
	}

	return
}

func (h *HtmlRender) Init(s RenderServer, funcs template.FuncMap) {
	h.root = template.New("REST_HTTP_ROOT")
	origin_funcs := template.FuncMap{"bower": RawHtml, "noescape": Noescape, "js": RawHtml, "css": RawHtml, "raw": RawHtml, "yield": RawHtml, "status": RawHtml, "slice": Slice, "mask": RawHtml, "repeat": Repeat}
	for k, v := range funcs {
		origin_funcs[k] = v
	}
	h.root.Funcs(origin_funcs)
	h.dir = s.WwwRoot()
	h.public = s.PublicDir()
	h.suffix = ".html"
	h.models = make(map[string]*template.Template)
	h.saveTemp = (s.Env() == "production")
	if h.saveTemp {
		h.initGlobalTemplate(h.root)
	}
}

func (h *HtmlRender) initTemplate(parent *template.Template, dir string, typ string) {
	parent.New("")
	if !h.saveTemp { //for debug
		log.Println("init template in ", h.dir, dir, "helper")
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

func (h *HtmlRender) initGlobalTemplate(parent *template.Template) {
	h.initTemplate(parent, ".", "global")
}

func (h *HtmlRender) initModelTemplate(parent *template.Template, dir string) {
	if dir != "" || dir != "." {
		h.initTemplate(parent, dir, "model")
	}
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
	if !h.saveTemp { //for debug
		log.Println("get template of ", name, file)
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
	layout   *template.Template
	yield    *template.Template
	css_file string
	js_file  string
	charset  string
}

func (h *HttpRenderInstance) Render(ctx *fasthttp.RequestCtx, data interface{}, status int, funcs template.FuncMap) error {
	ctx.SetContentType("text/html;; charset=" + h.charset)
	var mask_map = make(map[string]bool)

	funcMap := template.FuncMap{
		"yield": func() (template.HTML, error) {
			err := h.yield.Execute(ctx, data)
			if err != nil {
				log.Printf("%v%T", err, err)
			}
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
		"css": func() template.HTML {
			return template.HTML(`<link rel="stylesheet" type="text/css" href="` + h.css_file + `"></link>`)
		},
		"js": func() template.HTML {
			return template.HTML(`<script src="` + h.js_file + `"></script>`)
		},
		"sort": func(array []interface{}, by string) []interface{} {
			s := sorter{array, by}
			sort.Sort(&s)
			return array
		},
	}
	for k, v := range funcs {
		funcMap[k] = v
	}
	h.layout.Funcs(funcMap)
	h.yield.Funcs(funcMap)

	if h.layout != nil {
		return h.layout.Execute(ctx, data)
	} else if h.yield != nil {
		return h.yield.Execute(ctx, data)
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

// Htmlunquote returns unquoted html string.
func Noescape(src string) string {
	str, _ := url.QueryUnescape(src)
	return str
}
