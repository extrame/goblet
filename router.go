package goblet

import (
	"fmt"
	"net/http"
	"strings"

	"log/slog"

	"github.com/extrame/goblet/config"
	ge "github.com/extrame/goblet/error"
	"github.com/extrame/goblet/internal/ctrl"
	"github.com/extrame/goblet/internal/matcher"

	"github.com/pkg/errors"
)

type router struct {
	anchor *matcher.UrlMatcher
}

func (r *router) init() {
	r.anchor = matcher.Root(ctrl.NewStatic())
}

func (rou *router) route(s *Server, w http.ResponseWriter, r *http.Request) (err error) {
	defer func() {
		errorWrap(w)
	}()
	var anch *matcher.UrlMatcher
	var suffix_url string
	var main, format string
	var params map[string]string

	if r.URL.Path == "/" {
		anch, suffix_url, params = rou.anchor.Match("/index", 6)
		if !s.isSilence(r.URL.Path) {
			slog.Debug("routing /index", "path", r.URL.Path)
		}

	}

	context := &Context{
		s, r, w,
		nil, s.DB, suffix_url, format,
		"", nil, "default", nil, nil, nil, "", 200, false,
		nil, nil, nil,
		nil, false, params,
	}

	if s.nrPlugin != nil {
		s.nrPlugin.OnNewRequest(context)
	}

	if anch == nil {
		lastSlash := strings.LastIndex(r.URL.Path, "/")
		suff := strings.LastIndex(r.URL.Path, ".")
		if suff > 0 && suff < len(r.URL.Path) && suff > lastSlash {
			format = r.URL.Path[suff+1:]
			main = r.URL.Path[:suff]
		} else {
			main = r.URL.Path
		}
		var pathParams map[string]string
		// 尝试匹配，包括参数提取
		anch, suffix_url, pathParams = rou.anchor.Match(main, len(main))

		// 如果找到匹配，设置路径参数
		if anch != nil && len(pathParams) > 0 {
			for key, value := range pathParams {
				context.SetPathParam(key, value)
			}
			slog.Debug("matched with path params", "params", pathParams)
		}
		if !s.isSilence(r.URL.Path) {
			slog.Info("routing", "path", r.URL.Path)
		}
		if anch != nil {
			slog.Info("dynamic routing", "options", anch.Opt)
		}

	} else {
		format = "html"
	}

	if anch != nil {
		w.Header().Add("Cache-Control", "no-store,no-cache,must-revalidate,post-check=0,pre-check=0")
		w.Header().Add("Pragma", "no-cache")

		context.option = anch.Opt.(ctrl.Wrapper)
		context.suffix = suffix_url
		context.format = format

		if err = context.option.Parse(context); err == nil {
			context.checkResponse()
			if err = context.prepareRender(); err == nil {
				err = context.render()
			} else {
				err = errors.Wrap(err, fmt.Sprintf("[Original Response]%v", context.response))
			}
		}
		if s.Config.Basic.Env == config.DevelopEnv && err != nil && !s.isSilence(r.URL.Path) {
			slog.Info("Error in dynamic routing", "error", err)
		}

		return
	}
	return ge.NOSUCHROUTER("")
}

func (r *router) add(opt ctrl.Wrapper) {
	for _, v := range opt.GetRouting() {
		r.addRoute(v, opt)
	}
}

func (r *router) addRoute(path string, opt ctrl.Wrapper) {
	r.anchor.Add(path, opt)
}
