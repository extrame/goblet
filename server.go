package goblet

import (
	"flag"
	"fmt"
	toml "github.com/stvp/go-toml-config"
	"net/http"
	"path/filepath"
	"strings"
)

type Server struct {
	WwwRoot       *string
	PublicDir     *string
	ListenPort    *int
	IgnoreUrlCase *bool
	router        Router
	env           *string
	Renders       map[string]_Render
}

type Handler interface {
	Path() string
	Dir() string
}

type RestNewHander interface {
	New() (int, interface{})
}

type PageHandler interface {
	Page() (int, interface{})
}

func (s *Server) Organize(name string) {
	s.parseConfig(name)
	s.router.init()
	s.Renders = make(map[string]_Render)
	s.Renders["html"] = new(HtmlRender)
	s.Renders["html"].Init(s)
	s.Renders["json"] = new(JsonRender)
}

func (s *Server) Use(block interface{}) {
	cfg := PrepareOption(block)
	s.router.add(cfg)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ctx := s.router.route(w, r); ctx != nil {
		render := make(chan RenderInstance)
		response := make(chan *Response)
		go s.prepareRender(ctx, render)
		go ctx.handleData()
		for {
			select {
			case render_ready := <-render:
				ctx.renderInstance = render_ready
			case response_ready := <-response:
				ctx.response = response_ready
			}
			if ok := ctx.render(); ok {
				return
			}
		}
	} else {
		var path string
		if strings.HasSuffix(r.URL.Path, "/") {
			path = r.URL.Path + "index.html"
		} else {
			path = r.URL.Path
		}
		http.ServeFile(w, r, filepath.Join(*s.WwwRoot, *s.PublicDir, path))
	}
}

func (s *Server) prepareRender(c *Context, cha chan RenderInstance) {
	re := s.Renders[c.format]
	if re != nil {
		cha <- re.render(c.cfg)
	}
	cha <- nil
}

func (s *Server) parseConfig(name string) (err error) {
	path := flag.String("config", "./"+name+".conf", "设置配置文件的路径")
	s.WwwRoot = toml.String("basic.www_root", "./www")
	s.ListenPort = toml.Int("basic.port", 8080)
	s.PublicDir = toml.String("basic.public_dir", "public")
	s.IgnoreUrlCase = toml.Bool("basic.ignore_url_case", true)
	s.env = toml.String("basic.env", "development")
	flag.Parse()
	*path = filepath.FromSlash(*path)
	err = toml.Parse(*path)
	return
}

func (s *Server) RegisterRest(r interface{}, cfg RestOption) {
	// s.m.Group(cfg.Path, func(rou martini.Router) {
	// 	if new_handler, ok := r.(RestNewHander); ok {
	// 		rou.Get("/new", _RestRender{cfg, new_handler.New, "new"}.Wrap)
	// 	}
	// })
	return
}

func (s *Server) RegisterPage(r interface{}, cfg PageOption) {
	// if page_handler, ok := r.(PageHandler); ok {
	// 	s.m.Get(cfg.Path, _PageRender{cfg, page_handler.Page}.Wrap)
	// }
	return
}

func (s *Server) Run() {
	http.ListenAndServe(fmt.Sprintf(":%d", *s.ListenPort), s)
}
