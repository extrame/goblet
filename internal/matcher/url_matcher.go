package matcher

import (
	"strings"
)

type AnchorOp interface {
	MatchSuffix(string) bool
	String() string
}

// ---------------------anchors---------------
type UrlMatcher struct {
	loc      int
	char     string
	prefix   string
	branches []*UrlMatcher
	Opt      AnchorOp
}

func (a *UrlMatcher) Add(path string, opt AnchorOp) bool {
	if len(path) > a.loc {
		var full_stored_path = a.prefix + a.char
		if path[a.loc-len(a.prefix):a.loc+1] == full_stored_path {
			for _, v := range a.branches {
				if v.Add(path, opt) {
					return true
				}
			}
		}
		i := 0
		// for i := 0; i < len(full_stored_path); i++ {
		if path[a.loc+1-len(full_stored_path):a.loc+1-i] == full_stored_path[:len(full_stored_path)-i] {
			var branch *UrlMatcher
			if i != 0 {
				branch = &UrlMatcher{a.loc, a.char, strings.TrimPrefix(a.prefix, full_stored_path[:len(full_stored_path)-i]),
					a.branches, a.Opt}
				a.branches = []*UrlMatcher{branch}
			} else {
				if path[a.loc-len(a.prefix):] == full_stored_path {
					a.Opt = opt
					return true
				}
			}

			//add new b
			a.loc = a.loc - i
			branch = &UrlMatcher{len(path) - 1, path[len(path)-1:], path[a.loc+1 : len(path)-1], []*UrlMatcher{}, opt}
			a.branches = append(a.branches, branch)
			//change a
			a.char = full_stored_path[len(full_stored_path)-1-i : len(full_stored_path)-i]
			a.prefix = full_stored_path[:len(full_stored_path)-1-i]
			return true
		}
		// }
	}
	// else {
	// 	loc_begin_prefix := a.loc - len(a.prefix)
	// 	len_part_path := len(path) - loc_begin_prefix
	// 	for i := loc_begin_prefix + len_part_path - 1; i > loc_begin_prefix; i-- {
	// 		if path[loc_begin_prefix:i] == a.prefix[:i-loc_begin_prefix] {

	// 			//new branch for old
	// 			branch := &anchor{a.loc, a.char, a.prefix[i-loc_begin_prefix+1:], a.branches, a.opt}
	// 			a.branches = []*anchor{branch}

	// 			//change old
	// 			a.char = a.prefix[i-loc_begin_prefix-1 : i-loc_begin_prefix]
	// 			a.prefix = a.prefix[:i-loc_begin_prefix-1]
	// 			a.loc = i - 1

	// 			//new branch for new
	// 			branch = &anchor{len(path) - 1, path[len(path)-1:], path[a.loc : len(path)-1], []*anchor{}, opt}
	// 			a.branches = append(a.branches, branch)
	// 			return true
	// 		}
	// 	}
	// }
	return false
}

func (a *UrlMatcher) Match(path string, leng int) (*UrlMatcher, string) {
	if leng > a.loc && path[a.loc:a.loc+1] == a.char {
		if path[a.loc-len(a.prefix):a.loc] == a.prefix {
			for _, v := range a.branches {
				if res, suffix := v.Match(path, leng); res != nil {
					return res, suffix
				}
			}
			if a.Opt.MatchSuffix(path[a.loc+1:]) {
				return a, path[a.loc+1:]
			}
		}
	}
	return nil, ""
}

func Root(op AnchorOp) *UrlMatcher {
	return &UrlMatcher{0, "/", "", []*UrlMatcher{}, op}
}
