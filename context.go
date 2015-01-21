package goblet

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

var USERCOOKIENAME = "user"

type Context struct {
	Server         *Server
	Request        *http.Request
	writer         http.ResponseWriter
	option         BlockOption
	suffix         string
	format         string
	forceFormat    string
	tempRenders    []string
	method         string
	renderInstance RenderInstance
	response       interface{}
	responseMap    map[string]interface{}
	layout         string
	status_code    int
	already_writed bool
}

func (c *Context) handleData() {

}

func (c *Context) Writer() io.Writer {
	c.already_writed = true
	c.writer.WriteHeader(c.status_code)
	return c.writer
}

func (c *Context) SetHeader(key, value string) {
	c.writer.Header().Set(key, value)
}

func (c *Context) render() (err error) {
	if !c.already_writed {
		if c.renderInstance != nil {
			return c.renderInstance.render(c.writer, c.response, c.status_code)
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
			v := datas[i+1]
			c.responseMap[k] = v
		}
	}
}

func (c *Context) Respond(data interface{}) {
	switch data.(type) {
	case error:
		c.RespondWithStatus(data, http.StatusInternalServerError)
	default:
		c.RespondWithStatus(data, http.StatusOK)
	}
}

func (c *Context) RespondOK() {
	c.status_code = http.StatusOK
}

func (c *Context) RespondError(err error) {
	c.RespondWithStatus(err, http.StatusBadRequest)
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
	//test required format in allow list or not
	var format string
	if format, err = c.option.GetRender(c); err == nil {
		re := c.Server.Renders[format]
		if re != nil {
			c.renderInstance, err = re.render(c)
		}
	}
	return
}

func (c *Context) checkResponse() {
	if c.responseMap != nil {
		c.response = c.responseMap
	}
}

func (c *Context) Layout(l string) {
	c.layout = l
}

func (c *Context) getLayout() string {
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
