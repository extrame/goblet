package matcher

import (
	"fmt"
	"log/slog"
	"strings"
)

// hasPrefixAt 检查从指定位置开始的路径是否匹配给定前缀
// 避免了字符串切片分配，直接进行字节比较
func hasPrefixAt(path string, start int, prefix string) bool {
	if start < 0 || len(path)-start < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if path[start+i] != prefix[i] {
			return false
		}
	}
	return true
}

type AnchorOp interface {
	MatchSuffix(string) bool
	String() string
}

// ---------------------anchors---------------
type UrlMatcher struct {
	loc       int
	char      string
	prefix    string
	branches  []*UrlMatcher
	Opt       AnchorOp
	isParam   bool   // 是否为参数节点
	paramName string // 参数名称
}

type defaultAnchor struct{}

func (a *defaultAnchor) MatchSuffix(string) bool {
	return false
}

func (a *defaultAnchor) String() string {
	return "hidden default anchor"
}

func (a *UrlMatcher) String() string {
	return fmt.Sprintf("UrlMatcher(%s) on %s", a.prefix+a.char, a.Opt.String())
}

func (a *UrlMatcher) Add(path string, opt AnchorOp) bool {
	if len(path) > a.loc {
		var full_stored_path = a.prefix + a.char
		if path[a.loc-len(a.prefix):a.loc+1] == full_stored_path {
			for _, v := range a.branches {
				if v.Add(path[a.loc+1:], opt) {
					return true
				}
			}
		}
		if path[a.loc+1-len(full_stored_path):a.loc+1] == full_stored_path {
			if path[a.loc-len(a.prefix):] == full_stored_path {
				a.Opt = opt
				return true
			}
			return a.addSubPath(path[a.loc+1:], opt)
		}
	}
	return false
}

func (a *UrlMatcher) addSubPath(path string, opt AnchorOp) bool {
	var branch *UrlMatcher
	//add new b

	//find first location of :
	var tailBeforParam = strings.Index(path, ":")
	var paramName string
	var paramEnd int
	var hasParam = tailBeforParam > -1
	if tailBeforParam == -1 {
		tailBeforParam = len(path) - 1
	} else {
		paramEnd = strings.Index(path[tailBeforParam:], "/")
		if paramEnd == -1 {
			paramEnd = len(path) - tailBeforParam
		}
		if tailBeforParam == 0 && a.isParam {
			//如果是往一个参数节点添加子路径，应该跳过第一个参数
			return a.addSubPath(path[tailBeforParam+paramEnd:], opt)
		}
		paramName = path[tailBeforParam+1 : tailBeforParam+paramEnd]
		tailBeforParam = tailBeforParam - 1
	}
	var char = path[tailBeforParam : tailBeforParam+1]
	slog.Debug("add sub path", "path", path, "tailBeforParam", tailBeforParam, "char", char, "paramName", paramName)
	branch = &UrlMatcher{tailBeforParam,
		char, path[0:tailBeforParam],
		[]*UrlMatcher{}, nil, hasParam, paramName}
	if hasParam && tailBeforParam+paramEnd+1 < len(path) {
		branch.addSubPath(path[tailBeforParam+paramEnd+1:], opt)
	} else {
		branch.Opt = opt
	}
	a.branches = append(a.branches, branch)
	// change a
	// a.char = prefix[len(prefix)-1:]
	// a.prefix = prefix[:len(prefix)-1]
	return true
}

func (a *UrlMatcher) Match(path string, leng int, params ...map[string]string) (*UrlMatcher, string, map[string]string) {
	var paramsMap map[string]string
	if len(params) > 0 {
		paramsMap = params[0]
	} else {
		paramsMap = map[string]string{}
	}
	if leng > a.loc && path[a.loc:a.loc+1] == a.char {
		// 使用更高效的前缀比较方法，避免字符串切片分配
		if a.prefix == "" || (a.loc >= len(a.prefix) && hasPrefixAt(path, a.loc-len(a.prefix), a.prefix)) {
			var remainPath = path[a.loc+1:]
			if a.isParam {
				var nextSplit = strings.Index(remainPath, "/")
				if nextSplit == -1 {
					nextSplit = len(remainPath)
				}
				paramsMap[a.paramName] = remainPath[:nextSplit]
				remainPath = remainPath[nextSplit:]
			}
			// 首先尝试精确匹配
			for _, v := range a.branches {
				if res, suffix, params := v.Match(remainPath,
					len(remainPath), paramsMap); res != nil {
					return res, suffix, params
				}
			}
			if a.Opt.MatchSuffix(remainPath) {
				return a, remainPath, paramsMap
			}
		}
	}
	return nil, "", nil
}

func Root(op AnchorOp) *UrlMatcher {
	return &UrlMatcher{0, "/", "", []*UrlMatcher{}, op, false, ""}
}
