package goblet

import (
	"crypto/sha1"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/goblet/config"
	"github.com/extrame/goblet/error"
	"github.com/extrame/goblet/render"
	"github.com/go-xorm/xorm"
)

var NotImplemented = fmt.Errorf("this method is not implemented")

type Fn struct {
	Name string
	Fn   interface{}
}

type ControllerNeedInit interface {
	Init(*Server)
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
	HttpsEnable   *bool
	HttpsCertFile *string
	HttpsKey      *string
	HashSecret    *string
	dbEngine      *string
	dbUser        *string
	dbPwd         *string
	dbHost        *string
	dbName        *string
	version       *string
	dbPort        *int
	dbConTO       *int
	dbKaInterval  *int
	enDbCache     *bool
	cacheAmout    *int
	logFile       *string
	readTimeOut   *int
	writeTimeOut  *int
	Name          string
	plugins       map[string]Plugin
	funcs         []Fn
	initCtrl      []ControllerNeedInit
	pres          map[string]reflect.Value
	nrPlugin      onNewRequestPlugin
	saver         Saver
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

func (s *Server) Organize(name string, plugins []interface{}) {
	var err error
	var dbPwdPlugin DbPwdPlugin
	var dbUserPlugin dbUserNamePlugin
	s.Name = name
	for _, plugin := range plugins {
		if tp, ok := plugin.(Plugin); ok {
			typ := reflect.ValueOf(plugin).Type()
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			key := strings.ToLower(typ.Name())
			if s.plugins == nil {
				s.plugins = make(map[string]Plugin)
			}
			s.plugins[key] = tp
		}
		//bind the specials plugins
		if tp, ok := plugin.(DbPwdPlugin); ok {
			dbPwdPlugin = tp
		}
		if tp, ok := plugin.(dbUserNamePlugin); ok {
			dbUserPlugin = tp
		}
		if tp, ok := plugin.(onNewRequestPlugin); ok {
			s.nrPlugin = tp
		}
		if tp, ok := plugin.(Saver); ok {
			s.saver = tp
		}
	}
	if s.saver == nil {
		s.saver = new(LocalSaver)
	}
	s.pres = make(map[string]reflect.Value)
	if err = s.parseConfig(); err == nil {
		s.router.init()
		s.funcs = make([]Fn, 0)
		if dbPwdPlugin != nil {
			*(s.dbPwd) = dbPwdPlugin.SetPwd(*s.dbPwd)
		}
		if dbUserPlugin != nil {
			*(s.dbUser) = dbUserPlugin.SetName(*s.dbUser)
		}
		err = s.connectDB()
		if err == nil {
			if *s.env == config.DevelopEnv {
				log.Println("connect DB success")
				DB.ShowSQL(true)
			}
		} else if err != noneDbDriver {
			log.Fatalln("connect error:", err)
		}
	} else {
		log.Fatalln(err)
	}
	s.enableDbCache()
	for _, plugin := range s.plugins {
		plugin.Init(s)
	}
}

func (s *Server) connectDB() error {
	return newDB(*s.dbEngine, *s.dbUser, *s.dbPwd, *s.dbHost, *s.dbName, *s.dbPort, *s.dbConTO, *s.dbKaInterval)
}

func (s *Server) ControlBy(block interface{}) {
	cfg := PrepareOption(block)
	if bc, ok := block.(ControllerNeedInit); ok {
		s.initCtrl = append(s.initCtrl, bc)
	}
	s.router.add(cfg)
}

func (s *Server) caller() (string, string, error) {
	pc := make([]uintptr, 2) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[1])
	var caller_valid = regexp.MustCompile(`[\w]*\.\(\*([\w]+)\)\.([\w]+)`)
	matched := caller_valid.FindStringSubmatch(f.Name())
	if len(matched) == 3 {
		return matched[1], matched[2], nil
	}
	return "", "", errors.New("no matched caller")
}

func (s *Server) Pre(fn interface{}, conds ...string) {
	if c, _, err := s.caller(); err == nil {
		for _, m := range conds {
			key := strings.ToLower(c + "-" + m)
			s.pres[key] = reflect.ValueOf(fn)
		}
	}
}

func (s *Server) AddModel(models interface{}, syncs ...bool) {
	var err error

	var sync = true

	if len(syncs) > 0 {
		sync = syncs[0]
	}

	if sync {
		err = DB.Sync2(models)
		if err != nil {
			log.Fatalln("sync error:", err)
		}
	}
}

func (s *Server) Env() string {
	return *s.env
}

func (s *Server) Debug(fn func()) {
	if s.Env() == config.DevelopEnv {
		fn()
	}
}

func (s *Server) WwwRoot() string {
	if abs, err := filepath.Abs(*s.wwwRoot); err == nil {
		return abs
	} else {
		return *s.wwwRoot
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			s.wrapError(w, err, true)
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
		s.wrapError(w, err, false)
	}
}

func (s *Server) GetPlugin(key string) Plugin {
	return s.plugins[key]
}

func (s *Server) parseConfig() (err error) {
	path := flag.String("config", "./"+s.Name+".conf", "设置配置文件的路径")
	for key, plugin := range s.plugins {
		plugin.ParseConfig(key)
	}
	s.wwwRoot = toml.String("basic.www_root", "./www")
	s.ListenPort = toml.Int("basic.port", 8080)
	s.readTimeOut = toml.Int("basic.read_timeout", 30)
	s.writeTimeOut = toml.Int("basic.write_timeout", 30)

	s.publicDir = toml.String("basic.public_dir", "public")
	s.UploadsDir = toml.String("basic.uploads_dir", "./uploads")
	s.IgnoreUrlCase = toml.Bool("basic.ignore_url_case", true)
	s.HashSecret = toml.String("secret", "cX8Os0wfB6uCGZZSZHIi6rKsy7b0scE9")
	s.env = toml.String("basic.env", config.ProductEnv)
	s.dbEngine = toml.String("basic.db_engine", "mysql")
	s.enDbCache = toml.Bool("cache.enable", false)
	s.cacheAmout = toml.Int("cache.amount", 1000)
	s.logFile = toml.String("log.file", "")
	s.version = toml.String("basic.version", "")
	s.HttpsEnable = toml.Bool("basic.https", false)
	s.HttpsCertFile = toml.String("basic.https_certfile", "")
	s.HttpsKey = toml.String("basic.https_key", "")

	flag.Parse()
	*path = filepath.FromSlash(*path)
	err = toml.Parse(*path)
	if err == nil {
		s.initLog()
		s.dbHost = toml.String(*s.dbEngine+".host", s.Name)
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

func (s *Server) Run() error {
	s.Renders = make(map[string]render.Render)
	s.Renders["html"] = new(render.HtmlRender)
	var tempFuncMap = make(template.FuncMap)
	for _, bc := range s.initCtrl {
		bc.Init(s)
	}
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
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *s.ListenPort),
		Handler:      s,
		WriteTimeout: time.Second * time.Duration(*s.writeTimeOut),
		ReadTimeout:  time.Second * time.Duration(*s.readTimeOut),
	}
	var err error
	if !*s.HttpsEnable {
		err = srv.ListenAndServe()
	} else {
		err = srv.ListenAndServeTLS(*s.HttpsCertFile, *s.HttpsKey)
	}
	log.Println(err)
	return err
}
