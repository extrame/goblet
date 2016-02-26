package goblet

import (
	"bufio"
	"fmt"
	"github.com/extrame/goblet/render"
	"github.com/valyala/fasthttp"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

var USERCOOKIENAME = "user"

type Context struct {
	Server *Server
	ctx    *fasthttp.RequestCtx
	option BlockOption
	//默认请求类型：HTML
	suffix         string
	format         string
	forceFormat    string
	tempRenders    []string
	method         string
	renderInstance render.RenderInstance
	response       interface{}
	responseMap    map[string]interface{}
	layout         string
	status_code    int
	already_writed bool
	bower_stack    map[string]bool
	form_args      *fasthttp.Args
	form_parsed    bool
}

func (c *Context) handleData() {

}

func (c *Context) Writer() io.Writer {
	c.already_writed = true
	//TODO
	// c.writer.WriteHeader(c.status_code)
	return c.ctx
}

func (c *Context) Callback() string {
	return string(c.ctx.QueryArgs().Peek("callback"))
}

func (c *Context) Suffix() string {
	return c.suffix
}

func (c *Context) SetHeader(key, value string) {
	c.ctx.Response.Header.Set(key, value)
}

func (c *Context) IntFormValue(key string) int64 {
	s := c.StrFormValue(key)
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func (c *Context) StrFormValue(key string) string {
	bts := c.ctx.QueryArgs().Peek(key)
	if len(bts) == 0 {
		bts = c.ctx.PostArgs().Peek(key)
	}
	if len(bts) == 0 && c.parse_form() && c.form_args != nil {
		bts = c.form_args.Peek(key)
	}
	return string(bts)
}

func (c *Context) parse_form() bool {
	if c.form_args == nil && !c.form_parsed {
		if ct := string(c.ctx.Request.Header.Peek("Content-Type")); strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
			body := c.ctx.PostBody()
			args := new(fasthttp.Args)
			args.ParseBytes(body)
			c.form_args = args
		}
		c.form_parsed = true
	}
	return c.form_parsed
}

func (c *Context) render() (err error) {
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
			for i := 0; i < len(c.Server.funcs); i++ {
				var fn = c.Server.funcs[i].Fn
				funcMap[c.Server.funcs[i].Name] = func() interface{} {
					return fn(c)
				}
			}
			return c.renderInstance.Render(c.ctx, c.response, c.status_code, funcMap)
		} else {
			c.ctx.SetStatusCode(500)
			c.ctx.Write([]byte("Internal Error: No Render Allowed, please contact the admin"))
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
			v := autoHide(datas[i+1])
			c.responseMap[k] = v
		}
	}
}

func (c *Context) RespondReader(reader io.Reader) {
	bufio.NewWriter(c.ctx).ReadFrom(reader)
	c.already_writed = true
}

func (c *Context) Respond(data interface{}) {
	switch td := data.(type) {
	case error:
		c.RespondWithStatus(data, http.StatusInternalServerError)
	case []byte:
		c.format = "raw"
		c.Writer().Write(td)
	default:
		c.RespondWithStatus(data, http.StatusOK)
	}
}

func (c *Context) RespondOK() {
	c.status_code = http.StatusOK
}

func (c *Context) RespondError(err error) {
	c.RespondWithStatus(err.Error(), http.StatusBadRequest)
}

//Reset the context renders
func (c *Context) UseRender(render string) {
	c.forceFormat = render
}

//Allow some temporary render
func (c *Context) AllowRender(renders ...string) {
	c.tempRenders = renders
}

func (c *Context) RespondStatus(status int) {
	c.status_code = status
}

func (c *Context) RespondWithStatus(data interface{}, status int) {
	c.response = autoHide(data)
	c.status_code = status
}

func (c *Context) RespondWithRender(data interface{}, render string) {
	c.response = autoHide(data)
	c.method = render
}

func (c *Context) prepareRender() (err error) {
	if !c.already_writed {
		//test required format in allow list or not
		var format string
		if format, err = c.option.GetRender(c); err == nil {
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

func (c *Context) RemoteAddr() net.Addr {
	return c.ctx.RemoteAddr()
}

func (c *Context) FormValue(key string) string {
	return string(c.ctx.FormValue(key))
}

func (c *Context) RedirectTo(url string) {
	c.ctx.Response.Header.Set("Location", url)
	c.ctx.SetStatusCode(302)
	c.forceFormat = "raw"
}

//return the charset of the files
func (c *Context) CharSet() string {
	return *c.Server.charSet
}

///////////for renders/////////////
func (c *Context) BlockOptionType() string {
	switch c.option.(type) {
	case *RestBlockOption:
		return "Rest"
	case *GroupBlockOption:
		return "Group"
	case *_staticBlockOption:
		return "Static"
	case *HtmlBlockOption:
		return "Html"
	}
	log.Panic("err of block option type!!!")
	return ""
}

func (c *Context) QueryValue(key string) string {
	return string(c.ctx.URI().QueryString())
}

func (c *Context) Method() string {
	return c.method
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
