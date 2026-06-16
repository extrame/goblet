package ctrl

import (
	"fmt"
)

type Static struct {
	*Basic
}

func (c *Static) MatchSuffix(suffix string) bool {
	return true
}

func (c *Static) Parse(ctx Context) error {
	var suffix, suffixWithSlash = ctx.Suffix()
	if len(suffix) > 0 && suffixWithSlash {
		ctx.RenderAs(suffix)
	} else {
		ctx.RenderAs("index")
	}
	ctx.SetForceFormat("html", "default")
	return nil
}

func (h *Static) String() string {
	return fmt.Sprintf("Static(%s)", h.Name)
}

func NewStatic() Wrapper {
	return &Static{
		&Basic{},
	}
}
