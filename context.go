package goblet

import (
	"net/http"
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
}

func (c *Context) handleData() {

}

func (c *Context) render() error {
	return c.renderInstance.render(c.writer, c.response)
}

func (c *Context) prepareRender() {
	re := c.server.Renders[c.format]
	if re != nil {
		c.renderInstance = re.render(c)
	}
}

func (c *Context) RenderAs(name string) {
	c.option.UpdateRender(name, c)
}

func (c *Context) RestRedirectToRead(id string) {
	c.option.(*RestBlockOption).renderAsRead(id, c)
}

func (c *Context) GetLogin() (string, bool) {
	return c.GetLoginAs(USERCOOKIENAME)
}

func (c *Context) GetLoginAs(name string) (string, bool) {
	cookie, err := c.SignedCookie(name + "Id")
	if cookie != nil && err == nil {
		return cookie.Value, true
	}
	return "", false
}
