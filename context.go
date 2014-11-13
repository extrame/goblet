package goblet

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var USERCOOKIENAME = "user"

type Context struct {
	Server         *Server
	Request        *http.Request
	writer         http.ResponseWriter
	option         BlockOption
	suffix         string
	format         string
	method         string
	renderInstance RenderInstance
	response       interface{}
	responseMap    map[string]interface{}
	layout         string
	status_code    int
}

func (c *Context) handleData() {

}

func (c *Context) render() (err error) {
	return c.renderInstance.render(c.writer, c.response, c.status_code)
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

func (c *Context) prepareRender() {
	re := c.Server.Renders[c.format]
	if re != nil {
		c.renderInstance = re.render(c)
	}
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

func (c *Context) GetLoginId() (string, bool) {
	return c.GetLoginIdAs(USERCOOKIENAME)
}

func (c *Context) GetLoginIdAs(name string) (string, bool) {
	cookie, err := c.SignedCookie(name + "Id")
	if cookie != nil && err == nil {
		return cookie.Value, true
	}
	return "", false
}

func (c *Context) AddLoginId(id interface{}) {
	switch rid := id.(type) {
	case string:
		c.addLoginAs("user", rid)
	case int64:
		c.addLoginAs("user", strconv.FormatInt(rid, 10))
	}
}

func (c *Context) addLoginAs(name string, id string) {
	expire := time.Now().AddDate(0, 0, 1)
	cookie := new(http.Cookie)
	cookie.Name = name + "Id"
	cookie.Value = id
	cookie.Expires = expire
	cookie.RawExpires = expire.Format(time.UnixDate)
	fmt.Println(c.AddSignedCookie(cookie))
}
