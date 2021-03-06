package goblet

import (
	"crypto/sha1"
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

	"github.com/pkg/errors"
	"xorm.io/xorm/caches"

	"github.com/golang/glog"
	"github.com/sirupsen/logrus"

	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/goblet/config"
	ge "github.com/extrame/goblet/error"
	"github.com/extrame/goblet/render"
)

var NotImplemented = fmt.Errorf("this method is not implemented")

type Fn struct {
	Name string
	Fn   interface{}
}

type ControllerNeedInit interface {
	Init(*Server)
}

//Server 服务器类型
type Server struct {
	ConfigFile string

	wwwRoot         *string
	publicDir       *string
	UploadsDir      *string
	ListenPort      *int
	IgnoreUrlCase   *bool
	router          router
	env             *string
	Renders         map[string]render.Render
	HttpsEnable     *bool
	HttpsCertFile   *string
	HttpsKey        *string
	HashSecret      *string
	dbEngine        *string
	dbUser          *string
	dbPwd           *string
	dbHost          *string
	dbName          *string
	version         *string
	dbPort          *int
	dbConTO         *int
	dbKaInterval    *int
	enDbCache       *bool
	cacheAmout      *int
	logFile         *string
	readTimeOut     *int
	writeTimeOut    *int
	defaultType     *string
	enableKeepAlive *bool
	Name            string
	oldPlugins      map[string]Plugin
	plugins         map[string]NewPlugin
	funcs           []Fn
	initCtrl        []ControllerNeedInit
	pres            map[string][]reflect.Value
	nrPlugin        onNewRequestPlugin
	saver           Saver
	filler          map[string]FormFillFn
	multiFiller     map[string]MultiFormFillFn
	kv              KvDriver
	okFunc          func(*Context)
	errFunc         func(*Context, error, ...string)
	defaultRender   string
}

var defaultErrFunc = func(c *Context, err error, context ...string) {
	c.responseMap = nil
	msg := err.Error()
	if len(context) > 0 {
		msg = "[" + strings.Join(context, "|") + "]" + msg
	}
	c.RespondWithStatus(msg, http.StatusBadRequest)
}

func (s *Server) SetDefaultOk(fn func(*Context)) {
	s.okFunc = fn
}

func (s *Server) SetDefaultError(fn func(*Context, error, ...string)) {
	s.errFunc = fn
}

// type Handler interface {
// 	Path() string
// 	Dir() string
// }

// type RestNewHander interface {
// 	New() (int, interface{})
// }

// type PageHandler interface {
// 	Page() (int, interface{})
// }

//Organize 进行服务器环境的初始化配置，初始化所有plugin，对于plugin的所有操作，在Organize之后都可以视为已经初始化
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
			if s.oldPlugins == nil {
				s.oldPlugins = make(map[string]Plugin)
			}
			s.oldPlugins[key] = tp
		}
		if tp, ok := plugin.(NewPlugin); ok {
			typ := reflect.ValueOf(plugin).Type()
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			key := strings.ToLower(typ.Name())
			if s.plugins == nil {
				s.plugins = make(map[string]NewPlugin)
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
		if kv, ok := plugin.(KvDriver); ok {
			s.kv = kv
		}
		if ov, ok := plugin.(OkFuncSetter); ok {
			s.okFunc = ov.RespendOk
		}
		if ev, ok := plugin.(ErrFuncSetter); ok {
			s.errFunc = ev.RespondError
		}
		if rv, ok := plugin.(DefaultRenderSetter); ok {
			s.defaultRender = rv.DefaultRender()
		}
	}
	if s.saver == nil {
		s.saver = new(LocalSaver)
	}
	s.pres = make(map[string][]reflect.Value)
	s.filler = make(map[string]FormFillFn)
	s.multiFiller = make(map[string]MultiFormFillFn)
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
	for _, plugin := range s.oldPlugins {
		plugin.Init(s)
	}
	if s.errFunc == nil {
		s.errFunc = defaultErrFunc
	}
	if s.defaultRender == "" {
		s.defaultRender = "html"
	}
}

func (s *Server) connectDB() error {
	return newDB(*s.dbEngine, *s.dbUser, *s.dbPwd, *s.dbHost, *s.dbName, *s.dbPort, *s.dbConTO, *s.dbKaInterval)
}

//ControlBy
// Use Member of struct of type goblet.Router to redefine the path
func (s *Server) ControlBy(block interface{}) {
	cfg := s.prepareOption(block)
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
			if arr, ok := s.pres[key]; ok {
				s.pres[key] = append(arr, reflect.ValueOf(fn))
			} else {
				s.pres[key] = []reflect.Value{reflect.ValueOf(fn)}
			}
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
			glog.Fatalln("sync error:", err)
		}
	}
}

func (s *Server) Env() string {
	return *s.env
}

//Debug 当服务器环境为调试环境时，执行相应的匿名函数，用于编写调试环境专用的代码块
func (s *Server) Debug(fn func()) {
	if s.Env() == config.DevelopEnv {
		fn()
	}
}

func (s *Server) WwwRoot() string {
	if abs, err := filepath.Abs(*s.wwwRoot); err == nil {
		return abs
	}
	return *s.wwwRoot
}

func (s *Server) GetServerPathByCtrl(ctrl interface{}) []string {
	root := s.WwwRoot()
	cfg := s.prepareOption(ctrl)
	var paths = make([]string, len(cfg.GetRouting()))
	for i, r := range cfg.GetRouting() {
		paths[i] = filepath.Join(root, r)
	}
	return paths
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			s.wrapError(w, err, true)
		}
	}()
	if err := s.router.route(s, w, r); errors.Cause(err) == ge.NOSUCHROUTER {
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

//GetPlugin 获得对应名称的插件
func (s *Server) GetPlugin(key string) NewPlugin {
	return s.plugins[key]
}

func (s *Server) parseConfig() (err error) {
	flag.StringVar(&s.ConfigFile, "config", "./"+s.Name+".conf", "设置配置文件的路径")
	for key, plugin := range s.oldPlugins {
		glog.Errorln("使用旧版插件，请升级该插件到新版本")
		plugin.ParseConfig(key)
	}
	s.wwwRoot = toml.String("basic.www_root", "./www")
	s.ListenPort = toml.Int("basic.port", 8080)
	s.readTimeOut = toml.Int("basic.read_timeout", 30)
	s.writeTimeOut = toml.Int("basic.write_timeout", 30)

	s.publicDir = toml.String("basic.public_dir", "public")
	s.UploadsDir = toml.String("basic.uploads_dir", "./uploads")
	s.IgnoreUrlCase = toml.Bool("basic.ignore_url_case", true)
	s.HashSecret = toml.String("basic.secret", "a238974b2378c39021d23g43")
	s.env = toml.String("basic.env", config.ProductEnv)
	s.dbEngine = toml.String("basic.db_engine", "mysql")
	s.enDbCache = toml.Bool("cache.enable", false)
	s.cacheAmout = toml.Int("cache.amount", 1000)
	s.logFile = toml.String("log.file", "")
	s.version = toml.String("basic.version", "")
	s.HttpsEnable = toml.Bool("basic.https", false)
	s.HttpsCertFile = toml.String("basic.https_certfile", "")
	s.HttpsKey = toml.String("basic.https_key", "")
	s.defaultType = toml.String("basic.default_type", "")
	s.enableKeepAlive = toml.Bool("basic.keep_alive", true)

	flag.Parse()

	*s.env = strings.ToLower(*s.env)

	if *s.env != config.DevelopEnv && *s.env != config.ProductEnv && *s.env != config.OldProductEnv {
		glog.Fatalln("environment must be development or production")
	} else if *s.env == config.OldProductEnv {
		*s.env = config.ProductEnv
		fmt.Println("[Deprecatd]production environment must be set as 'production' instead of 'product'")
	}

	s.ConfigFile = filepath.FromSlash(s.ConfigFile)
	err = toml.Parse(s.ConfigFile)
	for _, plugin := range s.oldPlugins {
		plugin.Init(s)
	}
	for _, plugin := range s.plugins {
		if err = plugin.AddCfgAndInit(s); err != nil {
			glog.Fatalf("add plugin config error in (%T) with error (%s)", plugin, err)
		}
	}
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

//Hash 获得一个字符串的加密版本
func (s *Server) Hash(str string) string {
	hash := sha1.New()
	hash.Write([]byte(str))
	hash.Write([]byte(*s.HashSecret))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

//PublicDir 获得服务器对应的公共文件夹的地址
func (s *Server) PublicDir() string {
	return *s.publicDir
}

func (s *Server) enableDbCache() {
	if *s.enDbCache {
		cacher := caches.NewLRUCacher(caches.NewMemoryStore(), *s.cacheAmout)
		DB.SetDefaultCacher(cacher)
	}
}

//Run 运营一个服务器
func (s *Server) Run() error {
	if *s.version == "datetime" {
		*s.version = fmt.Sprintf("%d", time.Now().Unix())
	}
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
	s.Renders["xml"] = new(render.XmlRender)
	logrus.WithField("port", *s.ListenPort).Infoln("Listening")
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *s.ListenPort),
		Handler:      s,
		WriteTimeout: time.Second * time.Duration(*s.writeTimeOut),
		ReadTimeout:  time.Second * time.Duration(*s.readTimeOut),
	}
	if !(*s.enableKeepAlive) {
		srv.SetKeepAlivesEnabled(false)
	}
	var err error
	if !*s.HttpsEnable {
		err = srv.ListenAndServe()
	} else {
		err = srv.ListenAndServeTLS(*s.HttpsCertFile, *s.HttpsKey)
	}
	logrus.Println(err)
	return err
}
