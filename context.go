package goblet

import (
	"net/http"
)

type Context struct {
	req            *http.Request
	writer         http.ResponseWriter
	cfg            *BlockOption
	format         string
	renderInstance RenderInstance
	response       interface{}
}

func (c *Context) handleData() {

}

func (c *Context) render() bool {
	return c.renderInstance.render(c.writer, c.response)
}
