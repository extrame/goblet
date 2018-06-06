package goblet

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/extrame/goblet/render"
	"github.com/golang/glog"
)

var USERCOOKIENAME = "user"

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

type Context struct {
	Server  *Server
	request *http.Request
	writer  http.ResponseWriter
	option  BlockOption
	//默认请求类型：HTML
	suffix          string
	format          string
	forceFormat     string
	tempRenders     []string
	method          string
	renderInstance  render.RenderInstance
	response        interface{}
	responseMap     map[string]interface{}
	layout          string
	status_code     int
	already_writed  bool
	fill_bts        []byte
	bower_stack     map[string]bool
	infos           map[string]interface{}
	cookiesForWrite map[string]*http.Cookie
}

func (c *Context) handleData() {

}

func (cx *Context) GetRender() (render string, err error) {
	renders := cx.option.GetRender()
	if cx.forceFormat != "" {
		return cx.forceFormat, nil
	}
	if cx.format == "" {
		return renders[0], nil
	} else {
		for _, v := range renders {
			if v == cx.format {
				return v, nil
			}
		}
		if cx.tempRenders != nil {
			for _, v := range cx.tempRenders {
				if v == cx.format {
					return v, nil
				}
			}
		}
	}
	return "", RenderNotAllowd
}

func (c *Context) UseStandErrPage() bool {
	return c.option.ErrorRender() != "self"
}

func (c *Context) AddInfo(key string, value interface{}) {
	if c.infos == nil {
		c.infos = make(map[string]interface{})
	}
	c.infos[key] = value
}

func (c *Context) Writer() http.ResponseWriter {
	c.already_writed = true
	//TODO
	// c.writer.WriteHeader(c.status_code)
	return c.writer
}

func (c *Context) Callback() string {
	return c.request.FormValue("callback")
}

func (c *Context) Suffix() string {
	return c.suffix
}

func (c *Context) SetHeader(key, value string) {
	c.writer.Header().Add(key, value)
}

func (c *Context) IntFormValue(key string) int64 {
	str := c.request.FormValue(key)
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}

func (c *Context) StrFormValue(key string) string {
	return c.request.FormValue(key)
}

func (c *Context) EnableCache() {
	c.writer.Header().Del("Cache-Control")
	c.writer.Header().Del("Pragma")
}

func (c *Context) render() (err error) {
	for _, cookie := range c.cookiesForWrite {
		http.SetCookie(c.writer, cookie)
	}
	if !c.already_writed {
		if c.renderInstance != nil {
			funcMap := make(template.FuncMap)
			funcMap["bower"] = func(name string, version ...string) (template.HTML, error) {
				if c.bower_stack == nil {
					c.bower_stack = make(map[string]bool)
				}
				maps, err := c.Server.Bower(name, version...)
				res := ""
				for _, strs := range maps {
					if _, loaded := c.bower_stack[strs[0]]; !loaded {
						res += strs[1]
						c.bower_stack[strs[0]] = true
					}
				}
				return template.HTML(res), err
			}
			funcMap["version"] = func() string {
				return *c.Server.version
			}
			funcMap["extra_info"] = func(key string) interface{} {
				return c.infos[key]
			}
			for i := 0; i < len(c.Server.funcs); i++ {
				var fn = c.Server.funcs[i].Fn
				switch f := fn.(type) {
				case func(*Context) error:
					funcMap[c.Server.funcs[i].Name] = func() error {
						return f(c)
					}
				case func(*Context) interface{}:
					funcMap[c.Server.funcs[i].Name] = func() interface{} {
						return f(c)
					}
				case func(*Context, string) error:
					funcMap[c.Server.funcs[i].Name] = func(s string) error {
						return f(c, s)
					}
				case func(*Context, string) bool:
					funcMap[c.Server.funcs[i].Name] = func(s string) bool {
						return f(c, s)
					}
				case func(*Context, string, string) bool:
					funcMap[c.Server.funcs[i].Name] = func(s1, s2 string) bool {
						return f(c, s1, s2)
					}
				case func(*Context, string) interface{}:
					funcMap[c.Server.funcs[i].Name] = func(s string) interface{} {
						return f(c, s)
					}
				default:
					funcMap[c.Server.funcs[i].Name] = f
				}
			}
			return c.renderInstance.Render(c.writer, c.response, c.status_code, funcMap)
		} else {
			c.writer.WriteHeader(500)
			c.writer.Write([]byte("Internal Error: No Render Allowed, please contact the admin"))
		}
	}
	return nil
}

//Respond with multi data, data will tread as a key-value map
//for example:
//AddRespond("key1","value1",key2","value2")
//You can use AddRespond multi time in controller
func (c *Context) AddRespond(datas ...interface{}) {
	if len(datas) > 1 {
		if c.responseMap == nil {
			c.responseMap = make(map[string]interface{})
		}
		for i := 0; i < len(datas)/2; i++ {
			k := fmt.Sprintf("%s", datas[i])
			if c.option.AutoHidden() {
				v := autoHide(datas[i+1])
				c.responseMap[k] = v
			} else {
				c.responseMap[k] = datas[i+1]
			}
		}
	}
}

func (c *Context) RespondReader(reader io.Reader) {
	bufio.NewWriter(c.writer).ReadFrom(reader)
	c.already_writed = true
}

//Respond 向用户返回内容，三种数据会进行特别处理：
//error类型：标记状态为内部错误
//[]byte类型：使用raw进行内容渲染，即原样输出，不进行json等格式转化
//reader:使用raw进行内容渲染，即原样从输入中读取输出，不进行json等格式转化
func (c *Context) Respond(data interface{}) {
	switch td := data.(type) {
	case error:
		c.RespondWithStatus(data, http.StatusInternalServerError)
	case []byte:
		c.format = "raw"
		c.Writer().Write(td)
	case io.Reader:
		c.format = "raw"
		io.Copy(c.Writer(), td)
	default:
		c.RespondWithStatus(data, http.StatusOK)
	}
}

func (c *Context) RespondOK() {
	c.status_code = http.StatusOK
}

func (c *Context) RespondError(err error) {
	if c.Server.Env() == DevelopEnv {
		glog.Info("error is respond:", err)
	}
	c.responseMap = nil
	if err != nil {
		c.RespondWithStatus(err.Error(), http.StatusBadRequest)
	} else {
		c.RespondOK()
	}
}

//Reset the context renders
func (c *Context) UseRender(render string) {
	c.forceFormat = render
}

//AllowRender Allow some temporary render
func (c *Context) AllowRender(renders ...string) {
	c.tempRenders = renders
}

func (c *Context) RespondStatus(status int) {
	c.status_code = status
}

func (c *Context) RespondWithStatus(data interface{}, status int) {
	if c.option.AutoHidden() {
		c.response = autoHide(data)
	} else {
		c.response = data
	}
	c.status_code = status
}

func (c *Context) RespondWithRender(data interface{}, render string) {
	if c.option.AutoHidden() {
		c.response = autoHide(data)
	} else {
		c.response = data
	}
	c.method = render
}

func (c *Context) prepareRender() (err error) {
	if !c.already_writed {
		//test required format in allow list or not
		var format string
		if format, err = c.GetRender(); err == nil {
			re := c.Server.Renders[format]
			if re != nil {
				c.renderInstance, err = re.PrepareInstance(c)
			}
		}
	}
	return
}

func (c *Context) checkResponse() {
	if c.responseMap != nil {
		c.response = c.responseMap
	}
}

func (c *Context) SetLayout(l string) {
	c.layout = l
}

func (c *Context) Layout() string {
	if c.layout != "" {
		return c.layout
	} else {
		return c.option.Layout()
	}
}

func (c *Context) ResetDB() error {
	return c.Server.connectDB()
}

func (c *Context) RenderAs(name string) {
	c.method = name
}

func (c *Context) RestRedirectToRead(id interface{}) {
	switch rid := id.(type) {
	case string:
		c.option.(*RestBlockOption).renderAsRead(rid, c)
	case int64:
		c.option.(*RestBlockOption).renderAsRead(strconv.FormatInt(rid, 10), c)
	}
}

func (c *Context) RedirectTo(url string) {
	c.writer.Header().Set("Location", url)
	c.writer.WriteHeader(302)
	c.format = "raw"
}

///////////for renders/////////////
func (c *Context) BlockOptionType() string {
	switch c.option.(type) {
	case *RestBlockOption:
		return "Rest"
	case *groupBlockOption:
		return "Group"
	case *_staticBlockOption:
		return "Static"
	case *HtmlBlockOption:
		return "Html"
	}
	log.Panic("err of block option type!!!")
	return ""
}

//返回当前的Method
func (c *Context) Method() string {
	return c.method
}

//返回用户请求的Method
func (c *Context) ReqMethod() string {
	return c.request.Method
}

func (c *Context) Format() string {
	return c.format
}

func (c *Context) StatusCode() int {
	return c.status_code
}

func (c *Context) TemplatePath() string {
	return c.option.TemplatePath()
}

func (c *Context) String() string {
	return fmt.Sprintf("Context:{Type:%s,Method:%s}", c.BlockOptionType(), c.Method())
}

type RemoteAddr struct {
	str string
}

func (r *RemoteAddr) String() string {
	return r.str
}

func (r *RemoteAddr) Network() string {
	return "tcp"
}

func (c *Context) RemoteAddr() net.Addr {
	addr := new(RemoteAddr)
	addr.str = c.request.RemoteAddr
	return addr
}

func (c *Context) Form() url.Values {
	if c.request.Form == nil {
		c.request.ParseMultipartForm(defaultMaxMemory)
	}
	return c.request.Form
}

func (c *Context) FormValue(key string) string {
	return c.request.FormValue(key)
}

func (c *Context) Body() io.ReadCloser {
	return c.request.Body
}

func (c *Context) QueryString(key string) string {
	return c.request.URL.Query().Get(key)
}

func (c *Context) Referer() string {
	return c.request.Referer()
}

func (c *Context) PathToURL(path string) (*url.URL, error) {
	if abs, err := filepath.Abs(path); err == nil {
		if rel, err := filepath.Rel(c.Server.WwwRoot(), abs); err == nil {
			rel = strings.TrimPrefix(rel, "public/")
			return url.Parse("/" + rel)
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

//Version 返回用户配置的代码版本
func (c *Context) Version() string {
	return *c.Server.version
}

//ReqURL 返回用户请求的URL
func (c *Context) ReqURL() *url.URL {
	return c.request.URL
}

//ReqHeader 返回用户请求的Header
func (c *Context) ReqHeader() http.Header {
	return c.request.Header
}

//UserAgent 返回用户Agent
func (c *Context) UserAgent() string {
	return c.request.UserAgent()
}
