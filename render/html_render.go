package render

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/extrame/goblet/config"
	ge "github.com/extrame/goblet/error"
	"github.com/mvader/detect"
	"github.com/sirupsen/logrus"
)

var renderLock sync.Mutex

// var checkedAndStillNotExists = errors.New("checkedAndStillNotExists")

type HtmlRender struct {
	root      *template.Template
	dir       string
	models    *sync.Map
	pathRoot  *sync.Map
	suffix    string
	public    string
	saveTemp  bool
	notExists sync.Map
	delims    []string
}

func (h *HtmlRender) PrepareInstance(ctx RenderContext) (instance RenderInstance, err error) {
	var layout, yield *template.Template

	var model_root *template.Template

	var status_code = ctx.StatusCode()

	//for some refresh in test / develop env
	if !h.saveTemp {
		renderLock.Lock()
		model_root, _ = h.root.Clone()
		renderLock.Unlock()
		h.initGlobalTemplate(h.root)
		h.initModelHelperTemplate(model_root, ctx.TemplatePath())
	} else {
		if pR, ok := h.pathRoot.Load(ctx.TemplatePath()); ok {
			model_root = pR.(*template.Template)
		} else {
			model_root, _ = h.root.Clone()
			h.initModelHelperTemplate(model_root, ctx.TemplatePath())
			h.pathRoot.Store(ctx.TemplatePath(), model_root)
		}
	}

	path := ctx.TemplatePath() + "/" + ctx.Method()

	isMobile := detect.IsMobile(ctx.UserAgent())

	switch ctx.BlockOptionType() {

	case "Html":
		if isMobile {
			layout, err = h.getTemplates(model_root,
				"layout/"+ctx.Layout()+".mobile"+h.suffix, filepath.Join("layout", ctx.Layout()+".mobile"+h.suffix),
				"layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix),
			)
		} else {
			layout, err = h.getTemplates(model_root,
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
			layout, err = h.getTemplates(model_root,
				"layout/"+ctx.Layout()+".mobile"+h.suffix,
				filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+".mobile"+h.suffix),
				"layout/"+ctx.Layout()+h.suffix,
				filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+h.suffix),
				"layout/"+ctx.Layout()+".mobile"+h.suffix,
				filepath.Join("layout", ctx.Layout()+".mobile"+h.suffix),
				"layout/"+ctx.Layout()+h.suffix,
				filepath.Join("layout", ctx.Layout()+h.suffix),
			)
		} else {
			layout, err = h.getTemplates(model_root,
				"layout/"+ctx.Layout()+h.suffix, filepath.Join(ctx.TemplatePath(), "layout", ctx.Layout()+h.suffix),
				"layout/"+ctx.Layout()+h.suffix,
				filepath.Join("layout", ctx.Layout()+h.suffix),
			)
		}
	}

	//for static file 使得StaticRouter也可以使用404错误模版来渲染错误
	if ge.IsNoSuchRouter(err) && ctx.BlockOptionType() == "Static" {
		file := filepath.Join(h.dir, h.public, fmt.Sprintf("%s.%s", path, ctx.Format()))
		if _, err := os.Stat(file); os.IsNotExist(err) {
			status_code = 404
		}
	}

	if isMobile {
		yield, err = h.getTemplates(model_root,
			path+".mobile"+h.suffix, path+".mobile"+h.suffix,
			path+h.suffix, path+h.suffix)
	} else {
		yield, err = h.getTemplates(model_root, path+h.suffix, path+h.suffix)
	}

	if status_code >= 300 && ctx.UseStandErrPage() {
		if isMobile {
			layout, err = h.getTemplates(model_root,
				"layout/error.mobile."+h.suffix, filepath.Join("layout", "error.mobile."+h.suffix),
				"layout/error"+h.suffix, filepath.Join("layout", "error"+h.suffix),
				"layout/"+ctx.Layout()+".mobile"+h.suffix, filepath.Join("layout", ctx.Layout()+".mobile"+h.suffix),
				"layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix),
			)
		} else {
			layout, err = h.getTemplates(model_root,
				"layout/error"+h.suffix, filepath.Join("layout", "error"+h.suffix),
				"layout/"+ctx.Layout()+h.suffix, filepath.Join("layout", ctx.Layout()+h.suffix),
			)
		}
		var codePath = strconv.Itoa(status_code) + h.suffix
		yield, err = h.getTemplate(model_root, codePath, filepath.Join(path))
		if err != nil {
			logrus.Debugln("Find Err Code Fail, ", err)
			return nil, ge.NOSUCHROUTER(codePath)
		}
	}

	if err == nil {

		if h.delims != nil {
			yield.Delims(h.delims[0], h.delims[1])
			if layout != nil {
				layout.Delims(h.delims[0], h.delims[1])
			}
		}

		suffix := ""
		if v := ctx.Version(); v != "" {
			suffix = "?" + v
		}
		css := "/css/" + path + ".css"
		js := "/js/" + path + ".js"
		if isMobile {
			var mobile = "/css/" + path + ".mobile.css"
			if h.Exists(mobile) {
				css = mobile
			}
			mobile = "/js/" + path + ".mobile.js"
			if h.Exists(mobile) {
				js = mobile
			}
		}
		return &HttpRenderInstance{layout, yield, css + suffix, js + suffix}, nil
	} else {
		ce, ok := err.(*ge.Error)
		if ok {
			if ce.Code == ge.ERROR_CheckedAndStillNotExists {
				return nil, ge.NOSUCHROUTER(ce.Method)
			} else {
				logrus.Debugf("parse Template missing for %v", ctx)
				return nil, err
			}
		}
	}
	return nil, ge.NOSUCHROUTER("")
}

func (h *HtmlRender) Exists(file string) bool {
	info, err := os.Stat(filepath.Join(h.dir, filepath.FromSlash(file)))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
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
		"split":      Slice,
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
	h.delims = s.GetDelims()
	if h.saveTemp {
		h.initGlobalTemplate(h.root)
	}
}

func (h *HtmlRender) initHelperTemplate(parent *template.Template, dir string) {
	// parent.New("")
	logrus.Debug("init template in ", h.dir, dir, "helper")
	//scan for the helpers
	filepath.Walk(filepath.Join(h.dir, dir, "helper"), func(path string, info os.FileInfo, err error) error {
		if err == nil && (!info.IsDir()) && strings.HasSuffix(info.Name(), h.suffix) {
			name := strings.TrimSuffix(info.Name(), h.suffix)
			logrus.Infof("Parse helper:%s(%s)", name, path)
			e := parseFileWithName(parent, name, path)
			if e != nil {
				logrus.Infof("ERROR template.ParseFile: %v", e)
			}
		}
		return nil
	})
}

func (h *HtmlRender) initGlobalTemplate(parent *template.Template) {
	h.initHelperTemplate(parent, ".")
}

func (h *HtmlRender) initModelHelperTemplate(parent *template.Template, dir string) {
	if dir != "" && dir != "." {
		h.initHelperTemplate(parent, dir)
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
	file = filepath.FromSlash(file)
	var t *template.Template
	if t = root.Lookup(name); !h.saveTemp || t == nil {
		_, checked := h.notExists.Load(name)
		if !checked {
			//如果没检查过，打印日志，如果是原来检查不存在的，任何模式都跳过打日志
			logrus.Debugln("try to parse template of", name)
		} else if h.saveTemp {
			// saveTemp 产品模式，不再重复检查
			return nil, ge.CheckedAndStillNotExists(file)
		}
		err = parseFileWithName(root, name, filepath.Join(h.dir, file))
		if err == nil {
			t = root.Lookup(name)
		} else {
			if os.IsNotExist(err) {
				if !checked {
					logrus.Debugf("template for (%s) is missing", file)
					h.notExists.Store(name, true)
					return nil, ge.NOSUCHROUTER(file)
				}
				return nil, ge.CheckedAndStillNotExists(file)
			} else {
				return nil, err
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
}

func (h *HttpRenderInstance) Render(wr io.Writer, hwr HeadWriter, data interface{}, status int, funcs template.FuncMap) error {
	var mask_map = make(map[string]bool)

	funcMap := template.FuncMap{
		"yield": func() (html template.HTML, err error) {
			var temp *template.Template
			if temp, err = h.yield.Clone(); err == nil {
				err = temp.Execute(wr, data)
			}
			if err != nil {
				logrus.Error("[in yield]%v%T", err, err)
			}
			// return safe html here since we are rendering our own template
			return html, err
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
		temp, _ := h.layout.Clone()
		return temp.Execute(wr, data)
	} else if h.yield != nil {
		temp, _ := h.yield.Clone()
		return temp.Execute(wr, data)
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
	return parseBytesWithName(parent, name, s)
}

func parseBytesWithName(parent *template.Template, name string, content string) (err error) {
	renderLock.Lock()
	var tmpl *template.Template
	if name == parent.Name() || name == "" {
		tmpl = parent
	} else {
		tmpl = parent.New(name)
	}
	_, err = tmpl.Parse(content)
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
