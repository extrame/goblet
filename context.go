package goblet

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var USERCOOKIENAME = "user"

type Context struct {
	server         *Server
	req            *http.Request
	writer         http.ResponseWriter
	option         BlockOption
	suffix         string
	format         string
	method         string
	renderInstance RenderInstance
	response       interface{}
	layout         string
}

func (c *Context) handleData() {

}

func (c *Context) render() error {
	return c.renderInstance.render(c.writer, c.response)
}

func (c *Context) Respond(data interface{}) {
	c.response = autoHide(data)
}

func (c *Context) prepareRender() {
	re := c.server.Renders[c.format]
	if re != nil {
		c.renderInstance = re.render(c)
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
	c.option.UpdateRender(name, c)
}

func (c *Context) RestRedirectToRead(id interface{}) {
	switch rid := id.(type) {
	case string:
		c.option.(*RestBlockOption).renderAsRead(rid, c)
	case int64:
		c.option.(*RestBlockOption).renderAsRead(strconv.FormatInt(rid, 10), c)
	}
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
