package ctrl

import (
	"fmt"
	"strings"

	ge "github.com/extrame/goblet/v2/error"
)

const (
	REST_INDEX   = "index"   // GET /resources
	REST_SHOW    = "show"    // GET /resources/:id
	REST_NEW     = "new"     // GET /resources/new
	REST_CREATE  = "create"  // POST /resources
	REST_EDIT    = "edit"    // GET /resources/:id/edit
	REST_UPDATE  = "update"  // PATCH/PUT /resources/:id
	REST_DESTROY = "destroy" // DELETE /resources/:id
)

type Rest struct {
	*Basic
}

func (r *Rest) renderAsRead(id string, ctx Context) error {
	return r.callMethodForBlock("Read", ctx)
}

func (r *Rest) UpdateRender(obj string, ctx Context) {
	ctx.RenderAs(obj)
}

func (r *Rest) Parse(c Context) error {
	method := c.ReqMethod()
	if method == "OPTIONS" {
		c.SetHeader("Allow", "GET, POST, PATCH, PUT, DELETE")
		return nil
	}

	var id string
	var args []string
	var suffix = c.Suffix()
	if len(suffix) > 0 {
		id = suffix[1:]
		args = strings.SplitN(id, "/", 2)
		id = args[0]
	}

	switch {
	case id == "new" && method == "GET":
		c.RenderAs(REST_NEW)
		if args != nil && len(args) > 1 {
			suffix = args[1]
		}
		return r.Basic.callMethodForBlock("New", c)

	case id != "" && method == "GET":
		if nsuff := strings.TrimSuffix(suffix, "/edit"); nsuff != suffix {
			c.RenderAs(REST_EDIT)
			suffix = nsuff
			return r.Basic.callMethodForBlock("Edit", c)
		} else {
			c.RenderAs(REST_INDEX)
			er := r.Basic.callMethodForBlock("Index", c, true)

			if er != nil {
				c.RenderAs(REST_SHOW)
				return r.Basic.callMethodForBlock("Show", c)
			}
		}

	case id != "" && method == "DELETE":
		c.RenderAs(REST_DESTROY)
		return r.Basic.callMethodForBlock("Destroy", c)

	case id != "" && (method == "PATCH" || method == "PUT"):
		c.RenderAs(REST_UPDATE)
		return r.Basic.callMethodForBlock("Update", c)

	case id != "" && method == "POST":
		c.RenderAs(REST_CREATE)
		return r.Basic.callMethodForBlock("Create", c, true)

	case id == "" && method == "GET":
		c.RenderAs(REST_INDEX)
		return r.Basic.callMethodForBlock("Index", c)

	case id == "" && method == "POST":
		c.RenderAs(REST_CREATE)
		return r.Basic.callMethodForBlock("Create", c)
	}

	return ge.NOSUCHROUTER("")
}

func (r *Rest) handleData(c *Context) {

}

func (r *Rest) MatchSuffix(suffix string) bool {
	return len(suffix) == 0 || suffix[0:1] == "/"
}

func (h *Rest) String() string {
	return fmt.Sprintf("Rest(%s)", h.Name)
}
