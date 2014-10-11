package goblet

import (
	"net/http"
	"strings"
)

type Router struct {
	hockers map[string]*BlockOption
}

func (r *Router) init() {
	r.hockers = make(map[string]*BlockOption)
}

func (rou *Router) route(w http.ResponseWriter, r *http.Request) *Context {
	defer func() {
		ErrorWrap(w)
	}()
	var main, suffix string
	suff := strings.LastIndex(r.URL.Path, ".")
	if suff > 0 && suff < len(r.URL.Path) {
		suffix = r.URL.Path[suff+1:]
		main = r.URL.Path[:suff]
	} else {
		main = r.URL.Path
	}

	if opt, ok := rou.hockers[main]; ok {
		if suffix == "" && len(opt.render) > 0 {
			suffix = opt.render[0]
		} else {
			suffix = "html"
		}
		return &Context{r, w, opt, suffix, nil, nil}
	} else {
		return nil
	}

}

func (r *Router) add(opt *BlockOption) {
	for _, v := range opt.routing {
		r.addRoute(v, opt)
	}
}

func (r *Router) addRoute(path string, opt *BlockOption) {
	r.hockers[path] = opt
}
