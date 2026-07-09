package ctrl

import (
	"fmt"
	"log/slog"
	"strings"

	ge "github.com/extrame/goblet/v2/error"
)

type HttpMethod struct {
	*Basic
}

func (h *HttpMethod) String() string {
	return fmt.Sprintf("HttpMethod(%s)", h.Name)
}

func (h *HttpMethod) MatchSuffix(suffix string) bool {
	return len(suffix) == 0 || suffix[0:1] == "/"
}

func (h *HttpMethod) Parse(c Context) error {
	c.RenderAs(h.htmlRenderFileOrDir)
	method := c.ReqMethod()

	var suffix, suffixWithSlash = c.Suffix(true)

	if suffixWithSlash {
		var suffixWithMethod = strings.ToLower(method) + suffix
		matched, suffix, params := h.methods.Match(suffixWithMethod, len(suffixWithMethod))

		if matched != nil {
			if mc, ok := matched.Opt.(*MethodCaller); ok {
				mtd := mc.fn
				if mtd.IsValid() {
					c.SetSuffix(suffix)
					if h.tryPre(mc.String(), c) {
						results, typ := callMethod(mtd, c)
						return checkResult(results, typ, c)
					}
					return nil
				}
			} else {
				slog.Info("matched not method caller", "matched", matched)
			}
		}

		if params != nil {
			c.SetPathParams(params)
		}
	}

	return ge.NOSUCHROUTER("")

}
