package goblet

import (
	"fmt"

	"log/slog"

	"github.com/extrame/goblet/v2/internal/ctrl"
)

type Route byte
type Render byte
type Layout byte

// Controller which match full path according to RESTful rules, eg. a RestController with name Test will match /test and /test/1
type RestController byte

// Controller which match full path and path with any suffix, eg. a GroupController with name Test will match /test and /test/1 and /test/a/b/c
type GroupController byte

// Controller which match path and method
// eg. a HttpMethodController with name Test will match
// GET /test -> Test.Get
// POST /test -> Test.Post
// DELETE /test -> Test.Delete
// PUT /test -> Test.Put
// GET /test/config -> Test.GetConfig
// POST /test/info -> Test.PostInfo
type HttpMethodController byte

type ErrorRender byte
type AutoHide byte

// wrapController 将用户提交的controller进行适当封装，推断输入controller的类型和map方式
func (s *Server) wrapController(block interface{}) ctrl.Wrapper {

	basic, ignoreCase := ctrl.DetectOption(block, s)

	slog.Error("add routing", "ctrl", fmt.Sprintf("%T", block), "routing", basic.GetRouting())

	return ctrl.New(basic, block, ignoreCase)
}
