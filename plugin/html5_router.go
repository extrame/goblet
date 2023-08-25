package plugin

import (
	"path/filepath"

	"github.com/extrame/goblet"
	ge "github.com/extrame/goblet/error"
)

type Html5Router struct {
	goblet.Route  `/`
	goblet.Render `html=/`
	goblet.GroupController
	includes []string
	excludes []string
}

func (p *Html5Router) Get(ctx *goblet.Context) error {
	ctx.AllowRender(ctx.Format())

	for _, include := range p.includes {
		if matched, _ := filepath.Match(include, "/"+ctx.Suffix()); matched {
			goto matched
		}
	}

	for _, exclude := range p.excludes {
		if matched, _ := filepath.Match(exclude, "/"+ctx.Suffix()); matched {
			return ge.NOSUCHROUTER("")
		}
	}

matched:
	ctx.RenderAs("index")
	ctx.SetLayout("default")
	return nil
}

func Html5RoutePages(include []string, excluded ...[]string) *Html5Router {
	var router = &Html5Router{
		includes: include,
	}
	if len(excluded) > 0 {
		router.excludes = excluded[0]
	}
	return router
}
