package ctrl

import (
	"fmt"
	"strings"
)

type Html struct {
	*Basic
}

func (h *Html) String() string {
	return fmt.Sprintf("Html(%s)", h.Name)
}

func (h *Html) MatchSuffix(suffix string) bool {
	return len(suffix) == 0 || suffix[0:1] == "/"
}

func (h *Html) Parse(c Context) error {
	c.RenderAs(h.htmlRenderFileOrDir)
	method := c.ReqMethod()

	methodName := method[:1] + strings.ToLower(method[1:])
	return h.Basic.callMethodForBlock(methodName, c)
}
