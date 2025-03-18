package goblet

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/extrame/goblet/render"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
	showHidden      bool
}

func (c *Context) ShowHidden() {
	c.showHidden = true
}

func (c *Context) handleData() {

}

func (c *Context) DB() *gorm.DB {
	return c.Server.db
}

// GetRender,返回渲染类型,该返回需要判断是否允许相关渲染类型，如果不需要判断，请使用Format函数
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
	return "", fmt.Errorf("render (%s) is not allowed in %s", cx.format, cx.method)
}

func (c *Context) Format() string {
	return c.format
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

func (c *Context) GetInfo(key string) (interface{}, bool) {
	if val, ok := c.infos[key]; ok {
		return val, ok
	} else {
		return nil, false
	}
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

// EnableCache 手工启用缓存，默认Last-Modified为当前时间,Cache-Control为一年
func (c *Context) EnableCache() {
	c.EnableCacheFrom(time.Now())
}

// EnableCache 手工启用缓存，如果不传递时间参数，则默认Last-Modified为当前时间,Cache-Control为一年
// 如果传递一个时间参数，则Last-Modified为该时间，Cache-Control为一年，
// 如果传递两个时间参数，则Last-Modified为第一个时间，Cache-Control为第二个时间和第一个时间的差值
func (c *Context) EnableCacheFrom(lastModified time.Time, duration ...time.Duration) {
	c.writer.Header().Del("Pragma")
	var durationSeconds = 31536000
	if len(duration) > 0 {
		durationSeconds = int(duration[0].Seconds())
	}
	c.writer.Header().Set("Last-Modified", lastModified.Format(time.RFC1123))
	c.writer.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public, immutable", durationSeconds))
}

func (c *Context) render() (err error) {
	for _, cookie := range c.cookiesForWrite {
		http.SetCookie(c.writer, cookie)
	}
	if !c.already_writed {
		if c.renderInstance != nil {
			funcMap := make(template.FuncMap)
			funcMap["version"] = func() string {
				return c.Version()
			}
			funcMap["extra_info"] = func(key string) interface{} {
				return c.infos[key]
			}
			for i := 0; i < len(c.Server.funcs); i++ {
				var fn = c.Server.funcs[i].Fn

				rfn := reflect.ValueOf(fn)
				rfnT := rfn.Type()

				var nInTSlice = make([]reflect.Type, rfnT.NumIn())
				var num = 0
				var ctxIndex = -1

				for index := 0; index < rfnT.NumIn(); index++ {
					var in = rfnT.In(index)
					if in.Kind() == reflect.Ptr {
						ein := in.Elem()
						if ein.Name() == "Context" && ein.PkgPath() == "github.com/extrame/goblet" {
							ctxIndex = index
						}
					} else {
						nInTSlice[num] = in
						num++
					}
				}
				if ctxIndex >= 0 {
					nInTSlice = nInTSlice[:len(nInTSlice)-1]
				}
				var nOutTSlice = make([]reflect.Type, rfnT.NumOut())
				for index := 0; index < rfnT.NumOut(); index++ {
					nOutTSlice[index] = rfnT.Out(index)
				}
				var fT = reflect.FuncOf(nInTSlice, nOutTSlice, false)

				funcMap[c.Server.funcs[i].Name] = reflect.MakeFunc(fT, func(args []reflect.Value) []reflect.Value {
					var newArgs []reflect.Value
					var rc = reflect.ValueOf(c)
					if ctxIndex >= 0 {
						if ctxIndex == 0 {
							newArgs = append([]reflect.Value{rc}, args...)
						} else if ctxIndex >= len(args) {
							newArgs = append(args, rc)
						} else {
							newArgs = append(args[:ctxIndex], rc)
							newArgs = append(newArgs, args[ctxIndex:]...)
						}
						return rfn.Call(newArgs)
					}
					return rfn.Call(args)
				}).Interface()

				// switch f := fn.(type) {
				// case func(*Context) error:
				// 	funcMap[c.Server.funcs[i].Name] = func() error {
				// 		return f(c)
				// 	}
				// case func(*Context) interface{}:
				// 	funcMap[c.Server.funcs[i].Name] = func() interface{} {
				// 		return f(c)
				// 	}
				// case func(*Context) bool:
				// 	funcMap[c.Server.funcs[i].Name] = func() interface{} {
				// 		return f(c)
				// 	}
				// case func(*Context, string) error:
				// 	funcMap[c.Server.funcs[i].Name] = func(s string) error {
				// 		return f(c, s)
				// 	}
				// case func(*Context, string) bool:
				// 	funcMap[c.Server.funcs[i].Name] = func(s string) bool {
				// 		return f(c, s)
				// 	}
				// case func(*Context, string, string) bool:
				// 	funcMap[c.Server.funcs[i].Name] = func(s1, s2 string) bool {
				// 		return f(c, s1, s2)
				// 	}
				// case func(*Context, string) interface{}:
				// 	funcMap[c.Server.funcs[i].Name] = func(s string) interface{} {
				// 		return f(c, s)
				// 	}
				// default:
				// 	funcMap[c.Server.funcs[i].Name] = f
				// }
			}
			if c.ReqMethod() == "HEAD" {
				if hi, ok := c.renderInstance.(render.HeadRenderInstance); ok {
					return hi.HeadRender(&nullWriter{}, c.writer, c.response, c.status_code, funcMap)
				} else {
					return c.renderInstance.Render(&nullWriter{}, c.writer, c.response, c.status_code, funcMap)
				}
			}
			return c.renderInstance.Render(c.writer, c.writer, c.response, c.status_code, funcMap)
		} else {
			c.writer.WriteHeader(500)
			c.writer.Write([]byte("Internal Error: No Render Allowed, please contact the admin"))
		}
	}
	return nil
}

// Respond with multi data, data will tread as a key-value map
// for example:
// AddRespond("key1","value1",key2","value2")
// You can use AddRespond multi time in controller
func (c *Context) AddRespond(datas ...interface{}) {
	if len(datas) > 1 {
		if c.responseMap == nil {
			c.responseMap = make(map[string]interface{})
		}
		for i := 0; i < len(datas)/2; i++ {
			k := fmt.Sprintf("%s", datas[i])
			if c.showHidden || c.option.AutoHidden() {
				v := autoHide(datas[i+1], c)
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

// Respond 向用户返回内容，三种数据会进行特别处理：
// error类型：标记状态为内部错误
// []byte类型：使用raw进行内容渲染，即原样输出，不进行json等格式转化
// reader:使用raw进行内容渲染，即原样从输入中读取输出，不进行json等格式转化
func (c *Context) Respond(data interface{}) {
	switch td := data.(type) {
	case error:
		c.RespondWithStatus(data, http.StatusInternalServerError)
	case []byte, io.Reader, *File:
		c.UseRender("raw")
		c.RespondWithStatus(td, http.StatusOK)
	default:
		c.RespondWithStatus(data, http.StatusOK)
	}
}

func (c *Context) RespondField(data interface{}, fields ...string) {
	var objs = reflect.ValueOf(data)
	if objs.Type().Kind() == reflect.Ptr {
		objs = objs.Elem()
	}
	var result interface{}
	if objs.Type().Kind() == reflect.Slice || objs.Type().Kind() == reflect.Array {
		resultSlice := make([]map[string]interface{}, 0)
		for index := 0; index < objs.Len(); index++ {
			resultOne := getMapFromValue(objs.Index(index), fields)
			resultSlice = append(resultSlice, resultOne)
		}
		result = resultSlice
		c.Respond(result)
	} else {
		c.responseMap = getMapFromValue(objs, fields)
	}
}

func getMapFromValue(obj reflect.Value, fields []string) map[string]interface{} {
	resultOne := make(map[string]interface{})
	if obj.Type().Kind() == reflect.Ptr {
		obj = obj.Elem()
	}
	for _, field := range fields {
		var f = obj.FieldByName(field)
		if f.IsValid() {
			resultOne[field] = f.Interface()
		} else {
			resultOne[field] = nil
		}
	}
	return resultOne

}

func (c *Context) RespondOK() {
	if !c.already_writed {
		c.status_code = http.StatusOK
		if c.Server.okFunc != nil {
			c.Server.okFunc(c)
		}
	}
}

// RespondError 返回错误，如果错误为空，返回成功
func (c *Context) RespondError(err error, context ...string) {
	if c.Server.Env() == DevelopEnv {
		logrus.Info("error is respond:", err)
	}
	if err != nil {
		c.Server.errFunc(c, err, context...)
	} else {
		c.RespondOK()
	}
}

// Reset the context renders
func (c *Context) UseRender(render string) {
	c.forceFormat = render
	c.AllowRender("raw")
}

// AllowRender Allow some temporary render
func (c *Context) AllowRender(renders ...string) {
	c.tempRenders = renders
}

func (c *Context) RespondStatus(status int) {
	c.status_code = status
}

func (c *Context) RespondWithStatus(data interface{}, status int) {
	if c.showHidden || c.option.AutoHidden() {
		c.response = autoHide(data, c)
	} else {
		c.response = data
	}
	c.status_code = status
}

func (c *Context) RespondWithRender(data interface{}, render string) {
	if c.showHidden || c.option.AutoHidden() {
		c.response = autoHide(data, c)
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

// RenderAs 设置渲染的模型文件，注意和UseRender的区别，需要修改json/html等用UseRender
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
	c.writer.WriteHeader(301)
	c.format = "raw"
	c.already_writed = true
}

// /////////for renders/////////////
func (c *Context) BlockOptionType() string {
	return c.option.String()
}

// 返回当前的Method，json/html等
func (c *Context) Method() string {
	return c.method
}

// 返回用户请求的Method, GET/POST/HEAD等
func (c *Context) ReqMethod() string {
	return c.request.Method
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

// 该方法返回对端连接地址，返回有两种情况，如果头部没有X-Forwarded-For，表示是直接从客户端直接连接的。
// 此时，返回的addr，Network()方法返回“tcp"。如果头部有X-Forwarded-For，表示是从负载均衡跳转过来的。
// 此时，返回的addr，Network()方法返回“ip”
func (c *Context) RemoteAddr() net.Addr {
	addr := new(RemoteAddr)
	addr.str = c.request.RemoteAddr
	if remoteIP := c.request.Header.Get("X-Forwarded-For"); remoteIP != "" {
		return &net.IPAddr{
			IP: net.ParseIP(remoteIP),
		}
	}
	return addr
}

func (c *Context) RemoteIP() (net.IP, error) {
	if remoteIP := c.request.Header.Get("X-Forwarded-For"); remoteIP != "" {
		return net.ParseIP(remoteIP), nil
	} else {
		tcp, err := net.ResolveTCPAddr("tcp", c.request.RemoteAddr)
		if err == nil {
			return tcp.IP, nil
		} else {
			return nil, nil
		}
	}
}

// 返回对端地址，请注意，如果是负载均衡过来的请求，会直接返回负载均衡地址，不会重新获取客户端地址，如果要
// 获取这种情况下的客户端正式IP，请调用 RemoteAddr方法
func (c *Context) FromAddr() net.Addr {
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

func (c *Context) PostForm() url.Values {
	if c.request.PostForm == nil {
		c.request.ParseMultipartForm(defaultMaxMemory)
	}
	return c.request.PostForm
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

// Version 返回用户配置的代码版本
func (c *Context) Version() string {
	return c.Server.Basic.Version
}

// ReqURL 返回用户请求的URL
func (c *Context) ReqURL() *url.URL {
	return c.request.URL
}

// ReqHost 返回用户请求的host
func (c *Context) ReqHost() string {
	return c.request.Host
}

// ReqHeader 返回用户请求的Header
func (c *Context) ReqHeader() http.Header {
	return c.request.Header
}

// UserAgent 返回用户Agent
func (c *Context) UserAgent() string {
	return c.request.UserAgent()
}

// BasicAuth 返回Basic Auth
func (c *Context) BasicAuth() (string, string, bool) {
	return c.request.BasicAuth()
}
