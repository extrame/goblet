package goblet

import (
	"crypto/sha1"
	"flag"
	"fmt"
	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/goblet/error"
	"github.com/extrame/goblet/render"
	"github.com/go-xorm/xorm"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	DevelopEnv = "development"
	ProductEnv = "product"
)

var NotImplemented = fmt.Errorf("this method is not implemented")

type Fn struct {
	Name string
	Fn   func(*Context) interface{}
}

type Server struct {
	wwwRoot       *string
	publicDir     *string
	UploadsDir    *string
	ListenPort    *int
	IgnoreUrlCase *bool
	router        _Router
	env           *string
	Renders       map[string]render.Render
	HashSecret    *string
	dbEngine      *string
	dbUser        *string
	dbPwd         *string
	dbHost        *string
	dbName        *string
	dbPort        *int
	dbConTO       *int
	dbKaInterval  *int
	enDbCache     *bool
	cacheAmout    *int
	logFile       *string
	name          string
	plugins       []Plugin
	funcs         []Fn
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

func (s *Server) Organize(name string, plugins []Plugin) {
	var err error
	s.name = name
	s.plugins = plugins
	if err = s.parseConfig(); err == nil {
		s.router.init()
		s.funcs = make([]Fn, 0)
		if err = s.connectDB(); err == nil {
			if *s.env == DevelopEnv {
				DB.ShowSQL(true)
			}
		}
	} else {
		log.Fatalln(err)
	}
	s.enableDbCache()
	for _, plugin := range s.plugins {
		plugin.Init()
	}
}

func (s *Server) connectDB() error {
	return newDB(*s.dbEngine, *s.dbUser, *s.dbPwd, *s.dbHost, *s.dbName, *s.dbPort, *s.dbConTO, *s.dbKaInterval)
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

func (s *Server) Env() string {
	return *s.env
}

func (s *Server) WwwRoot() string {
	return *s.wwwRoot
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			WrapError(w, err, true)
		}
	}()
	if err := s.router.route(s, w, r); err == ge.NOSUCHROUTER {
		var path string
		if strings.HasSuffix(r.URL.Path, "/") {
			path = r.URL.Path + "index.html"
		} else {
			path = r.URL.Path
		}
		http.ServeFile(w, r, filepath.Join(*s.wwwRoot, s.PublicDir(), path))
	} else if err != nil {
		WrapError(w, err, false)
	}
}

func (s *Server) parseConfig() (err error) {
	path := flag.String("config", "./"+s.name+".conf", "设置配置文件的路径")
	for _, plugin := range s.plugins {
		plugin.ParseConfig()
	}
	s.wwwRoot = toml.String("basic.www_root", "./www")
	s.ListenPort = toml.Int("basic.port", 8080)
	s.publicDir = toml.String("basic.public_dir", "public")
	s.UploadsDir = toml.String("basic.uploads_dir", "./uploads")
	s.IgnoreUrlCase = toml.Bool("basic.ignore_url_case", true)
	s.HashSecret = toml.String("secret", "cX8Os0wfB6uCGZZSZHIi6rKsy7b0scE9")
	s.env = toml.String("basic.env", "production")
	s.dbEngine = toml.String("basic.db_engine", "mysql")
	s.enDbCache = toml.Bool("cache.enable", false)
	s.cacheAmout = toml.Int("cache.amount", 1000)
	s.logFile = toml.String("log.file", "")
	flag.Parse()
	*path = filepath.FromSlash(*path)
	err = toml.Parse(*path)
	if err == nil {
		s.initLog()
		s.dbHost = toml.String(*s.dbEngine+".host", s.name)
		s.dbUser = toml.String(*s.dbEngine+".user", "")
		s.dbPwd = toml.String(*s.dbEngine+".password", "")
		s.dbName = toml.String(*s.dbEngine+".name", "")
		s.dbPort = toml.Int(*s.dbEngine+".port", 3306)
		s.dbConTO = toml.Int(*s.dbEngine+".connect_timeout", 30)
		s.dbKaInterval = toml.Int(*s.dbEngine+".ka_interval", 0)
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

func (s *Server) PublicDir() string {
	return *s.publicDir
}

func (s *Server) enableDbCache() {
	if *s.enDbCache {
		cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), *s.cacheAmout)
		DB.SetDefaultCacher(cacher)
	}
}

func (s *Server) Run() {
	s.Renders = make(map[string]render.Render)
	s.Renders["html"] = new(render.HtmlRender)
	var tempFuncMap = make(template.FuncMap)
	for _, v := range s.funcs {
		tempFunc := func() int {
			return 0
		}
		tempFuncMap[v.Name] = tempFunc
	}
	s.Renders["html"].Init(s, tempFuncMap)
	s.Renders["json"] = new(render.JsonRender)
	s.Renders["raw"] = new(render.RawRender)
	log.Println("Listen at ", fmt.Sprintf(":%d", *s.ListenPort))
	http.ListenAndServe(fmt.Sprintf(":%d", *s.ListenPort), s)
}
