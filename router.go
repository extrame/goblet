package goblet

import (
	"github.com/extrame/goblet/error"
	"log"
	"net/http"
	"strings"
)

type _Router struct {
	anchor *Anchor
}

func (r *_Router) init() {
	r.anchor = &Anchor{0, "/", "", []*Anchor{}, &_staticBlockOption{}}
}

func (rou *_Router) route(s *Server, w http.ResponseWriter, r *http.Request) (err error) {
	defer func() {
		ErrorWrap(w)
	}()
	var anch *Anchor
	var suffix_url string
	var main, suffix string

	if r.URL.Path == "/" {
		anch, suffix_url = rou.anchor.match("/index", 6)
		log.Printf("routing /index\n", r.URL.Path)
	}

	if anch == nil {
		suff := strings.LastIndex(r.URL.Path, ".")
		if suff > 0 && suff < len(r.URL.Path) {
			suffix = r.URL.Path[suff+1:]
			main = r.URL.Path[:suff]
		} else {
			main = r.URL.Path
		}
		anch, suffix_url = rou.anchor.match(main, len(main))
		log.Printf("routing matched %s\n", r.URL.Path)
	} else {
		suffix = "html"
	}

	if anch != nil {
		context := &Context{s, r, w, anch.opt, suffix_url, suffix, "", nil, "default", nil, nil, nil, "", 200, false, nil}
		if err = anch.opt.Parse(context); err == nil {
			context.checkResponse()
			if err = context.prepareRender(); err == nil {
				err = context.render()
			}
		}
		if *s.env == DevelopEnv {
			log.Println("Err in Dynamic :", err)
		}
		return
	}
	return ge.NOSUCHROUTER
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
		var full_stored_path = a.prefix + a.char
		if path[a.loc-len(a.prefix):a.loc+1] == full_stored_path {
			for _, v := range a.branches {
				if v.add(path, opt) {
					return true
				}
			}
		}
		for i := 0; i < len(full_stored_path); i++ {
			if path[a.loc+1-len(full_stored_path):a.loc+1-i] == full_stored_path[:len(full_stored_path)-i] {
				var branch *Anchor
				if i != 0 {
					branch = &Anchor{a.loc, a.char, strings.TrimPrefix(a.prefix, full_stored_path[:len(full_stored_path)-i]), a.branches, a.opt}
					a.branches = []*Anchor{branch}
				} else {
					log.Println(":", path[a.loc-len(a.prefix):], full_stored_path)
					if path[a.loc-len(a.prefix):] == full_stored_path {
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
	} else {
		loc_begin_prefix := a.loc - len(a.prefix)
		len_part_path := len(path) - loc_begin_prefix
		for i := loc_begin_prefix + len_part_path - 1; i > loc_begin_prefix; i-- {
			if path[loc_begin_prefix:i] == a.prefix[:i-loc_begin_prefix] {

				log.Println(i, loc_begin_prefix, a.prefix[i-loc_begin_prefix+1:])
				log.Println(path[loc_begin_prefix:i], a.prefix[:i-loc_begin_prefix-1], a.prefix[i-loc_begin_prefix-1:i-loc_begin_prefix])

				//new branch for old
				branch := &Anchor{a.loc, a.char, a.prefix[i-loc_begin_prefix+1:], a.branches, a.opt}
				a.branches = []*Anchor{branch}

				//change old
				a.char = a.prefix[i-loc_begin_prefix-1 : i-loc_begin_prefix]
				a.prefix = a.prefix[:i-loc_begin_prefix-1]
				a.loc = i - 1

				//new branch for new
				branch = &Anchor{len(path) - 1, path[len(path)-1:], path[a.loc : len(path)-1], []*Anchor{}, opt}
				a.branches = append(a.branches, branch)
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
