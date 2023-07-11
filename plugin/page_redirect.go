package plugin

import (
	"regexp"

	"github.com/extrame/goblet"
	ge "github.com/extrame/goblet/error"
)

type PageRedirector struct {
	goblet.Route  `/`
	goblet.Render `html`
	goblet.GroupController
	matcher             *regexp.Regexp
	target              string
	withOriginalAsQuery bool
}

func (p *PageRedirector) Get(ctx *goblet.Context) error {
	ctx.AllowRender(ctx.Format())
	if p.matcher.MatchString("/" + ctx.Suffix()) {
		var target = p.target
		if p.withOriginalAsQuery {
			target += "?original=" + ctx.Suffix()
		}
		ctx.RedirectTo(target)
	}
	return ge.NOSUCHROUTER("")
}

//PageRedirect Create a page redirector match the matcher and redirect to target, if withOriginalAsQuery is true, the original url will be append to the target url's query part as target?original=original
func PageRedirect(matcher *regexp.Regexp, target string, withOriginalAsQuery bool) *PageRedirector {
	return &PageRedirector{
		matcher:             matcher,
		target:              target,
		withOriginalAsQuery: withOriginalAsQuery,
	}
}
