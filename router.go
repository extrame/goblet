package goblet

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type _Router struct {
	anchor *Anchor
}

var NOSUCHROUTER = fmt.Errorf("no such router")

func (r *_Router) init() {
	r.anchor = &Anchor{0, "/", "", []*Anchor{}, &StaticBlockOption{}}
}

func (rou *_Router) route(s *Server, w http.ResponseWriter, r *http.Request) (err error) {
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
		suffix = "html"
	}

	log.Println("routing " + r.URL.Path)

	anch, suffix_url := rou.anchor.match(main, len(main))

	if anch != nil {
		context := &Context{s, r, w, anch.opt, suffix_url, suffix, "default", nil, nil, "", 200}
		if err = anch.opt.Parse(context); err == nil {
			context.prepareRender()
			err = context.render()
		}
		return
	}
	return NOSUCHROUTER
}

func (r *_Router) add(opt BlockOption) {
	for _, v := range opt.GetRouting() {
		r.addRoute(v, opt)
	}
}

func (r *_Router) addRoute(path string, opt BlockOption) {
	r.anchor.add(path, opt)
}

//---------------------anchors---------------
type Anchor struct {
	loc      int
	char     string
	prefix   string
	branches []*Anchor
	opt      BlockOption
}

func (a *Anchor) add(path string, opt BlockOption) bool {
	if len(path) > a.loc {
		if path[a.loc-len(a.prefix):a.loc+1] == a.prefix+a.char {
			for _, v := range a.branches {
				if v.add(path, opt) {
					return true
				}
			}
		}
		var full_stored_path = a.prefix + a.char
		for i := 0; i < len(full_stored_path); i++ {
			if path[a.loc+1-len(full_stored_path):a.loc+1-i] == full_stored_path[:len(full_stored_path)-i] {

				var branch *Anchor
				if i != 0 {
					branch = &Anchor{a.loc, a.char, strings.TrimPrefix(a.prefix, full_stored_path[:len(full_stored_path)-i]), a.branches, opt}
					a.branches = []*Anchor{branch}
				} else {
					if path == full_stored_path {
						a.opt = opt
						return true
					}
				}

				//add new b
				a.loc = a.loc - i
				branch = &Anchor{len(path) - 1, path[len(path)-1:], path[a.loc+1 : len(path)-1], []*Anchor{}, opt}
				a.branches = append(a.branches, branch)
				//change a
				a.char = full_stored_path[len(full_stored_path)-1-i : len(full_stored_path)-i]
				a.prefix = full_stored_path[:len(full_stored_path)-1-i]
				return true
			}
		}

	}
	return false
}

func (a *Anchor) match(path string, leng int) (*Anchor, string) {
	if leng > a.loc && path[a.loc:a.loc+1] == a.char {
		if path[a.loc-len(a.prefix):a.loc] == a.prefix {
			for _, v := range a.branches {
				if res, suffix := v.match(path, leng); res != nil {
					return res, suffix
				}
			}
			if a.opt.MatchSuffix(path[a.loc+1:]) {
				return a, path[a.loc+1:]
			}
		}
	}
	return nil, ""
}
