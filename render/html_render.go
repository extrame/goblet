package render

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/extrame/goblet/config"
	"github.com/extrame/goblet/error"
	"github.com/golang/glog"
	"github.com/mvader/detect"
)

var renderLock sync.Mutex

type HtmlRender struct {
	root     *template.Template
	dir      string
	models   *sync.Map
	pathRoot *sync.Map
	suffix   string
	public   string
	saveTemp bool
}

func (h *HtmlRender) PrepareInstance(ctx RenderContext) (instance RenderInstance, err error) {
	var layout, yield *template.Template

	var root *template.Template

	var status_code = ctx.StatusCode()

	//for some refresh in test / develop env
	if !h.saveTemp {
		root, _ = h.root.Clone()
		h.initGlobalTemplate(root)
		h.initModelHelperTemplate(root, ctx.TemplatePath())
	} else {
		if pR, ok := h.pathRoot.Load(ctx.TemplatePath()); ok {
			root = pR.(*template.Template)
		} else {
			root, _ = h.root.Clone()
			h.initModelHelperTemplate(root, ctx.TemplatePath())
		}
	}

	path := ctx.TemplatePath() + "/" + ctx.Method()

	isMobile := detect.IsMobile(ctx.UserAgent())

	switch ctx.BlockOptionType() {

	case "Html":
		if isMobile {
			layout, err = h.getTemplates(root,
				"layout/"+ctx.Layout()+".mobile"+h.suffix, filepath.Join("layout", ctx.Layout()+".mobile"+h.suffix),
				"layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix),
			)
		} else {
			layout, err = h.getTemplates(root,
				"layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix),
			)
		}
		path = ctx.Method()
	case "Static":
		ctx.EnableCache()
		path = ctx.Method()
		fallthrough
	case "Rest", "Group":
		if isMobile {
			layout, err = h.getTemplates(root,
				ctx.TemplatePath()+"/layout/"+ctx.Layout()+".mobile"+h.suffix, filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+".mobile"+h.suffix),
				ctx.TemplatePath()+"/layout/"+ctx.Layout()+h.suffix, filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+h.suffix),
				"layout/"+ctx.Layout()+".mobile"+h.suffix, filepath.Join("layout", ctx.Layout()+".mobile"+h.suffix),
				"layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix),
			)
		} else {
			layout, err = h.getTemplates(root,
				ctx.TemplatePath()+"/layout/"+ctx.Layout()+h.suffix, filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+h.suffix),
				"layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix),
			)
		}
	}

	//for static file
	if err == ge.NOSUCHROUTER && ctx.BlockOptionType() == "Static" {
		file := filepath.Join(h.dir, h.public, fmt.Sprintf("%s.%s", path, ctx.Format()))
		if _, err := os.Stat(file); os.IsNotExist(err) {
			status_code = 404
		}
	}

	if err == nil {
		if isMobile {
			yield, err = h.getTemplates(root,
				path+".mobile"+h.suffix, path+".mobile"+h.suffix,
				path+h.suffix, path+h.suffix)
		} else {
			yield, err = h.getTemplates(root, path+h.suffix, path+h.suffix)
		}
	}

	if status_code >= 300 && ctx.UseStandErrPage() {
		if isMobile {
			layout, err = h.getTemplates(root,
				"layout/error.mobile."+h.suffix, filepath.Join("layout", "error.mobile."+h.suffix),
				"layout/error"+h.suffix, filepath.Join("layout", "error"+h.suffix),
				"layout/"+ctx.Layout()+".mobile"+h.suffix, filepath.Join("layout", ctx.Layout()+".mobile"+h.suffix),
				"layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix),
			)
		} else {
			layout, err = h.getTemplates(root,
				"layout/"+ctx.Layout()+".mobile"+h.suffix, filepath.Join("layout", ctx.Layout()+".mobile"+h.suffix),
				"layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix),
			)
		}
		yield, err = h.getTemplate(root, strconv.Itoa(status_code)+h.suffix, filepath.Join(strconv.Itoa(status_code)+h.suffix))
		if err != nil {
			glog.Infoln("Find Err Code Fail, ", err)
		}
	}

	if err == nil {
		suffix := ""
		if v := ctx.Version(); v != "" {
			suffix = "?" + v
		}
		mobile := ""
		if isMobile {
			mobile = ".mobile"
		}
		return &HttpRenderInstance{layout, yield, "/css/" + path + mobile + ".css" + suffix, "/js/" + path + mobile + ".js" + suffix}, nil
	} else {
		glog.Errorf("parse Template error for %v", ctx)
	}

	return
}

func (h *HtmlRender) Init(s RenderServer, funcs template.FuncMap) {
	h.root = template.New("REST_HTTP_ROOT")
	origin_funcs := template.FuncMap{
		"extra_info": RawHtml,
		"version":    RawHtml,
		"bower":      RawHtml,
		"noescape":   Noescape,
		"js":         RawHtml,
		"css":        RawHtml,
		"raw":        RawHtml,
		"yield":      RawHtml,
		"status":     RawHtml,
		"slice":      Slice,
		"last":       Last,
		"first":      First,
		"mask":       RawHtml,
		"repeat":     Repeat}
	for k, v := range funcs {
		origin_funcs[k] = v
	}
	h.root.Funcs(origin_funcs)
	h.dir = s.WwwRoot()
	h.public = s.PublicDir()
	h.suffix = ".html"
	h.models = new(sync.Map)
	h.pathRoot = new(sync.Map)
	h.saveTemp = (s.Env() == config.ProductEnv)
	if h.saveTemp {
		h.initGlobalTemplate(h.root)
	}
}

func (h *HtmlRender) initHelperTemplate(parent *template.Template, dir string, typ string) {
	// parent.New("")
	if !h.saveTemp { //for debug
		glog.Infoln("init template in ", h.dir, dir, "helper")
	}
	//scan for the helpers
	filepath.Walk(filepath.Join(h.dir, dir, "helper"), func(path string, info os.FileInfo, err error) error {
		if err == nil && (!info.IsDir()) && strings.HasSuffix(info.Name(), h.suffix) {
			name := strings.TrimSuffix(info.Name(), h.suffix)
			glog.Infof("Parse helper:%s(%s)", typ+"/"+name, path)
			e := parseFileWithName(parent, typ+"/"+name, path)
			if e != nil {
				glog.Infof("ERROR template.ParseFile: %v", e)
			}
		}
		return nil
	})
}

func (h *HtmlRender) initGlobalTemplate(parent *template.Template) {
	h.initHelperTemplate(parent, ".", "global")
}

func (h *HtmlRender) initModelHelperTemplate(parent *template.Template, dir string) {
	if dir != "" || dir != "." {
		h.initHelperTemplate(parent, dir, "model")
	}
}

func (h *HtmlRender) getTemplates(root *template.Template, args ...string) (temp *template.Template, err error) {
	if len(args)%2 == 0 {
		for index := 0; index < len(args)/2; index++ {
			name := args[index*2]
			file := args[index*2+1]
			if temp, err = h.getTemplate(root, name, file); err == nil {
				return
			}
		}
		return
	}
	return nil, errors.New("Input length of args is odd")
}

func (h *HtmlRender) getTemplate(root *template.Template, args ...string) (*template.Template, error) {
	var name, file string
	var err error
	if len(args) == 1 {
		name = args[0]
		file = args[0]
	} else {
		name = args[1]
		file = args[1]
	}
	if !h.saveTemp { //for debug
		glog.Infoln("get template of", name, file)
	}
	file = filepath.FromSlash(file)
	var t interface{}
	var ok bool
	if t, ok = h.models.Load(name); !ok {
		if h.saveTemp {
			glog.Infoln("try to parse template of", name)
		}
		var cloned_rest_model *template.Template
		cloned_rest_model, err = root.Clone()

		if err == nil {

			err = parseFileWithName(cloned_rest_model, name, filepath.Join(h.dir, file))
			if err == nil {
				t = cloned_rest_model.Lookup(name)
				if h.saveTemp {
					h.models.Store(name, t)
				}
			} else {
				if os.IsNotExist(err) {
					glog.Errorf("template for (%s) is missing", file)
					return nil, ge.NOSUCHROUTER
				} else {
					return nil, err
				}
			}
		}
	}
	return t.(*template.Template), nil
}

type HttpRenderInstance struct {
	layout   *template.Template
	yield    *template.Template
	css_file string
	js_file  string
}

func (h *HttpRenderInstance) Render(wr http.ResponseWriter, data interface{}, status int, funcs template.FuncMap) error {
	var mask_map = make(map[string]bool)

	funcMap := template.FuncMap{
		"yield": func() (template.HTML, error) {
			err := h.yield.Execute(wr, data)
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
	renderLock.Lock()
	_, err = tmpl.Parse(s)
	renderLock.Unlock()
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

func Last(obj interface{}) interface{} {
	slice := reflect.ValueOf(obj)

	return slice.Index(slice.Len() - 1).Interface()
}

func First(obj interface{}) interface{} {
	slice := reflect.ValueOf(obj)

	return slice.Index(0).Interface()
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
