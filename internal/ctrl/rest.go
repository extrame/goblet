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
	return r.Basic.callMethodForBlock(ctx, CallParams{
		MethodName: "Read",
		Params:     []string{id},
	})
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

	var id string = c.PopSuffix()
	var suffix, _ = c.Suffix()

	switch {
	case id == "new" && method == "GET":
		c.RenderAs(REST_NEW)
		return r.Basic.callMethodForBlock(c, CallParams{
			MethodName: "New",
		})

	case id != "" && method == "GET":
		if nsuff := strings.TrimSuffix(suffix, "/edit"); nsuff != suffix {
			c.RenderAs(REST_EDIT)
			suffix = nsuff
			return r.Basic.callMethodForBlock(c, CallParams{
				MethodName: "Edit",
			})
		} else {
			c.RenderAs(REST_INDEX)
			//只测试id作为method的第二个部分的函数是否存在，存在则调用
			er := r.Basic.callMethodForBlock(c, CallParams{
				MethodName:    "Index",
				SubMethodName: id,
				OnlyWithParam: true,
			})

			if er != nil {
				c.RenderAs(REST_SHOW)
				return r.Basic.callMethodForBlock(c, CallParams{
					MethodName: "Show",
					Params:     []string{id},
				})
			}
		}

	case id != "" && method == "DELETE":
		c.RenderAs(REST_DESTROY)
		return r.Basic.callMethodForBlock(c, CallParams{
			MethodName: "Destroy",
			Params:     []string{id},
		})

	case id != "" && (method == "PATCH" || method == "PUT"):
		c.RenderAs(REST_UPDATE)
		return r.Basic.callMethodForBlock(c, CallParams{
			MethodName: "Update",
			Params:     []string{id},
		})

	case id != "" && method == "POST":
		c.RenderAs(REST_CREATE)
		return r.Basic.callMethodForBlock(c, CallParams{
			MethodName:    "Create",
			OnlyWithParam: true,
			Params:        []string{id},
		})

	case id == "" && method == "GET":
		c.RenderAs(REST_INDEX)
		return r.Basic.callMethodForBlock(c, CallParams{
			MethodName: "Index",
		})

	case id == "" && method == "POST":
		c.RenderAs(REST_CREATE)
		return r.Basic.callMethodForBlock(c, CallParams{
			MethodName: "Create",
		})
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
