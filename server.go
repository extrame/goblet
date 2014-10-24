package goblet

import (
	"crypto/sha1"
	"flag"
	"fmt"
	toml "github.com/extrame/go-toml-config"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

type Server struct {
	WwwRoot       *string
	PublicDir     *string
	ListenPort    *int
	IgnoreUrlCase *bool
	router        _Router
	env           *string
	Renders       map[string]_Render
	HashSecret    *string
	dbEngine      *string
	dbUser        *string
	dbPwd         *string
	dbHost        *string
	dbName        *string
	dbPort        *int
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

func (s *Server) Organize(name string, opts ...Option) {
	var err error
	if err = s.parseConfig(name); err == nil {
		var opt Option
		opt.overlay(opts)
		s.router.init()
		s.Renders = make(map[string]_Render)
		s.Renders["html"] = new(HtmlRender)
		s.Renders["html"].Init(s)
		s.Renders["json"] = new(JsonRender)
		s.Renders["raw"] = new(RawRender)
		if err = newDB(*s.dbEngine, *s.dbUser, *s.dbPwd, *s.dbHost, *s.dbName, *s.dbPort); err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Fatalln(err)
	}

}

func (s *Server) ControlBy(block interface{}) {
	cfg := PrepareOption(block)
	s.router.add(cfg)
}

func (s *Server) AddModel(models ...interface{}) {
	var err error

	err = DB.Sync2(models...)
	if err != nil {
		log.Fatalln(err)
	}

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
	s.dbEngine = toml.String("basic.db_engine", "mysql")
	flag.Parse()
	*path = filepath.FromSlash(*path)
	err = toml.Parse(*path)
	if err == nil {
		s.dbHost = toml.String(*s.dbEngine+".host", "")
		s.dbUser = toml.String(*s.dbEngine+".user", "")
		s.dbPwd = toml.String(*s.dbEngine+".password", "")
		s.dbName = toml.String(*s.dbEngine+".name", "")
		s.dbPort = toml.Int(*s.dbEngine+".port", 3306)
		err = toml.Load()
	}
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
