package goblet

import (
	"crypto/sha1"
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
	HashSecret    *string
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
	if err := s.router.route(s, w, r); err != nil {
		var path string
		if strings.HasSuffix(r.URL.Path, "/") {
			path = r.URL.Path + "index.html"
		} else {
			path = r.URL.Path
		}
		http.ServeFile(w, r, filepath.Join(*s.WwwRoot, *s.PublicDir, path))
	}
}

func (s *Server) parseConfig(name string) (err error) {
	path := flag.String("config", "./"+name+".conf", "设置配置文件的路径")
	s.WwwRoot = toml.String("basic.www_root", "./www")
	s.ListenPort = toml.Int("basic.port", 8080)
	s.PublicDir = toml.String("basic.public_dir", "public")
	s.IgnoreUrlCase = toml.Bool("basic.ignore_url_case", true)
	s.HashSecret = toml.String("secret", "cX8Os0wfB6uCGZZSZHIi6rKsy7b0scE9")
	s.env = toml.String("basic.env", "development")
	flag.Parse()
	*path = filepath.FromSlash(*path)
	err = toml.Parse(*path)
	return
}

func (s *Server) Hash(str string) string {

	hash := sha1.New()
	hash.Write([]byte(str))
	hash.Write([]byte(*s.HashSecret))
	return fmt.Sprintf("%x", hash.Sum(nil))

}

func (s *Server) Run() {
	http.ListenAndServe(fmt.Sprintf(":%d", *s.ListenPort), s)
}
