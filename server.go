package goblet

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/extrame/goblet/config"
	ge "github.com/extrame/goblet/error"
	"github.com/extrame/goblet/render"
	yaml "gopkg.in/yaml.v3"
)

var NotImplemented = fmt.Errorf("this method is not implemented")

type Fn struct {
	Name string
	Fn   interface{}
}

type ControllerNeedInit interface {
	Init(*Server)
}

type ControllerNeedInitAndReturnError interface {
	Init(*Server) error
}

// Server 服务器类型
type Server struct {
	ConfigFile string

	Basic   config.Basic
	Cache   config.Cache
	Log     config.Log
	Db      config.Db
	router  router
	Renders map[string]render.Render

	Name          string
	plugins       map[string]NewPlugin
	funcs         []Fn
	initCtrl      []ControllerNeedInit
	initCtrlNew   []ControllerNeedInitAndReturnError
	pres          map[string][]reflect.Value
	nrPlugin      onNewRequestPlugin
	saver         Saver
	filler        map[string]FormFillFn
	multiFiller   map[string]MultiFormFillFn
	kv            KvDriver
	okFunc        func(*Context)
	errFunc       func(*Context, error, ...string)
	defaultRender string
	cfg           *yaml.Node
	cfgFileSuffix string
	silenceUrls   map[string]bool
	loginSaver    LoginInfoStorer
	configer      Configer
	delims        []string
	db            *gorm.DB
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

// Organize 进行服务器环境的初始化配置，初始化所有plugin，对于plugin的所有操作，在Organize之后都可以视为已经初始化
func (s *Server) Organize(name string, plugins []interface{}) {
	var err error
	var dbPwdPlugin DbPwdPlugin
	var dbUserPlugin dbUserNamePlugin
	s.Name = name
	s.Renders = make(map[string]render.Render)
	for _, plugin := range plugins {
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
		if tp, ok := plugin.(ChangeSuffixOfConfig); ok {
			s.cfgFileSuffix = tp.GetConfigSuffix()
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
		if sv, ok := plugin.(SilenceUrlSetter); ok {
			s.silenceUrls = sv.SetSilenceUrls()
		}
		if lv, ok := plugin.(LoginInfoStorer); ok {
			s.loginSaver = lv
		}
		if lv, ok := plugin.(Configer); ok {
			s.configer = lv
		}
		if dv, ok := plugin.(DelimSetter); ok {
			var delimis = dv.SetDelim()
			s.delims = delimis[:]
		}
		if rv, ok := plugin.(render.Render); ok {
			s.Renders[rv.Type()] = rv
		}
	}
	if s.saver == nil {
		s.saver = new(LocalSaver)
	}
	if s.configer == nil {
		s.configer = new(YamlConfiger)
	}
	if s.loginSaver == nil {
		s.loginSaver = new(CookieLoginInfoStorer)
	}
	s.pres = make(map[string][]reflect.Value)
	s.filler = make(map[string]FormFillFn)
	s.multiFiller = make(map[string]MultiFormFillFn)
	if err = s.parseConfig(); err == nil {
		s.router.init()
		s.funcs = make([]Fn, 0)
		if dbPwdPlugin != nil {
			s.Db.Pwd = dbPwdPlugin.SetPwd(s.Db.Pwd)
		}
		if dbUserPlugin != nil {
			s.Db.User = dbUserPlugin.SetName(s.Db.User)
		}
		err = s.connectDB()
		if err == nil {
			if s.Basic.Env == config.DevelopEnv {
				s.db.Config.Logger = logger.Default.LogMode(logger.Info)
			}
		} else if err != config.NoDbDriver {
			slog.Error("Failed to connect to database", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Error("Failed to read config file", "error", err)
		os.Exit(1)
	}
	s.enableDbCache()
	if s.errFunc == nil {
		s.errFunc = defaultErrFunc
	}
	if s.defaultRender == "" {
		s.defaultRender = "html"
	}
}

func (s *Server) isSilence(u string) bool {
	si, ok := s.silenceUrls[u]
	return ok && si
}

func (s *Server) connectDB() (err error) {
	dialectorCreator, ok := dialectorCreators[s.Basic.DbEngine]
	if !ok {
		return config.NoDbDriver
	}
	s.db, err = s.Db.New(s.Basic.DbEngine, dialectorCreator)
	return err
}

// ControlBy
// Use Member of struct of type goblet.Router to redefine the path
func (s *Server) ControlBy(block interface{}) {
	cfg := s.prepareOption(block)
	if bc, ok := block.(ControllerNeedInit); ok {
		s.initCtrl = append(s.initCtrl, bc)
	}
	if bc, ok := block.(ControllerNeedInitAndReturnError); ok {
		s.initCtrlNew = append(s.initCtrlNew, bc)
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
		err = s.db.AutoMigrate(models)
		if err != nil {
			slog.Error("Failed to sync database model", "error", err)
			os.Exit(1)
		}
	}
}

func (s *Server) Env() string {
	return s.Basic.Env
}

// Debug 当服务器环境为调试环境时，执行相应的匿名函数，用于编写调试环境专用的代码块
func (s *Server) Debug(fn func()) {
	if s.Env() == config.DevelopEnv {
		fn()
	}
}

func (s *Server) WwwRoot() string {
	if abs, err := filepath.Abs(s.Basic.WwwRoot); err == nil {
		return abs
	}
	return s.Basic.WwwRoot
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
	err := s.router.route(s, w, r)
	err = errors.Cause(err)
	if geE, ok := err.(*ge.Error); ok && geE.Code == ge.ERROR_NOSUCHROUTER {
		var path string
		if geE.Method != "" {
			//dynamic return a method which should used as static render
			slog.Debug("Using static file from dynamic response",
				"method", geE.Method)
			file := filepath.Join(s.Basic.WwwRoot, s.PublicDir(), geE.Method)
			if _, err := os.Stat(file); !os.IsNotExist(err) {
				s.ServeFile(w, r, filepath.Join(s.Basic.WwwRoot, s.PublicDir(), geE.Method))
				return
			}
		}
		if strings.HasSuffix(r.URL.Path, "/") {
			path = r.URL.Path + "index.html"
		} else {
			path = r.URL.Path
		}
		s.ServeFile(w, r, filepath.Join(s.Basic.WwwRoot, s.PublicDir(), path))
	} else if err != nil {
		s.wrapError(w, err, false)
	}
}

func (s *Server) ServeFile(w http.ResponseWriter, r *http.Request, file string) {
	//if not index.html, set cache-control to 1 year
	if filepath.Base(file) != "index.html" {
		w.Header().Del("Pragma")
		w.Header().Set("Cache-Control", "max-age=31536000")
	}

	http.ServeFile(w, r, file)
}

// GetPlugin 获得对应名称的插件
func (s *Server) GetPlugin(key string) NewPlugin {
	return s.plugins[key]
}

func (s *Server) SetConfigSuffix(suffix string) {
	s.cfgFileSuffix = suffix
}

// getCfg 从 YAML 配置中获取指定键名的配置节点
// 返回一个可以用于解码的 yaml.Node
func (s *Server) getCfg(key string) *yaml.Node {
	// 确保配置已加载
	if s.cfg == nil || s.cfg.Kind != yaml.DocumentNode {
		return &yaml.Node{Kind: yaml.ScalarNode}
	}

	// 获取根节点（通常是一个映射节点）
	root := s.cfg.Content[0]
	if root.Kind != yaml.MappingNode {
		return &yaml.Node{Kind: yaml.ScalarNode}
	}

	// 遍历映射节点的键值对
	for i := 0; i < len(root.Content); i += 2 {
		// Content 数组中，偶数索引是键，奇数索引是值
		if root.Content[i].Value == key {
			return root.Content[i+1]
		}
	}

	// 如果没有找到对应的键，返回一个空的标量节点
	return &yaml.Node{Kind: yaml.ScalarNode}
}

func (s *Server) parseConfig() (err error) {
	reader, err := s.configer.GetConfigSource(s)
	if err == nil {
		s.initLog()
		s.cfg = new(yaml.Node)
		err = yaml.NewDecoder(reader).Decode(s.cfg)
		if err == nil {
			if err = s.getCfg("basic").Decode(&s.Basic); err == nil {
				s.Db.Name = s.Name
				if err = s.getCfg(s.Basic.DbEngine).Decode(&s.Db); err == nil {
					if s.Db.Host == "" {
						s.Db.Host = s.Name
					}
					if err = s.getCfg("cache").Decode(&s.Cache); err == nil {
						s.getCfg("log").Decode(&s.Log)
					}
				}
			}
		}
	}

	if err != nil {
		return err
	}

	if s.Basic.Env == "" {
		s.Basic.Env = config.DevelopEnv
		slog.Info("Environment not set, using development as default")
	}

	if s.Basic.DbEngine == "" {
		s.Basic.DbEngine = "none"
	}

	if s.Basic.Port == 0 {
		s.Basic.Port = 8080
	}

	if s.Basic.Env != config.DevelopEnv && s.Basic.Env != config.ProductEnv && s.Basic.Env != config.OldProductEnv {
		slog.Error("Invalid environment setting",
			"message", "Environment must be development or production",
			"valid_values", []string{"development", "production"})
		os.Exit(1)
	} else if s.Basic.Env == config.OldProductEnv {
		s.Basic.Env = config.ProductEnv
		slog.Warn("Deprecated environment value",
			"message", "Use 'production' instead of 'product'")
	}
	// Note: slog level should be set globally when initializing the application
	for _, plugin := range s.plugins {
		if err = plugin.AddCfgAndInit(s); err != nil {
			slog.Error("Failed to initialize plugin",
				"plugin_type", fmt.Sprintf("%T", plugin),
				"error", err)
			os.Exit(1)
		}
	}
	return
}

// Hash 获得一个字符串的加密版本
func (s *Server) Hash(str string) string {
	hash := sha1.New()
	hash.Write([]byte(str))
	hash.Write([]byte(s.Basic.HashSecret))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// PublicDir 获得服务器对应的公共文件夹的地址
func (s *Server) PublicDir() string {
	return s.Basic.PublicDir
}

func (s *Server) enableDbCache() {
	// GORM doesn't have built-in caching system
	// If caching is needed, consider using a separate caching layer
	// such as Redis or implementing a custom caching middleware
	if s.Cache.Enable {
		slog.Info("GORM database caching not available",
			"message", "Consider using a separate caching layer such as Redis")
	}
}

func (s *Server) GetDelims() []string {
	return s.delims
}

// AddConfig 从服务器配置中获取指定名称的配置节点，并反序列化到提供的对象中
// name: 配置节点的名称
// obj: 接收配置的对象指针
// 返回错误信息，如果反序列化成功则返回 nil
func (s *Server) AddConfig(name string, obj interface{}) error {
	// 检查参数
	if name == "" {
		return fmt.Errorf("config name cannot be empty")
	}
	if obj == nil {
		return fmt.Errorf("target object cannot be nil")
	}

	// 检查配置是否已加载
	if s.cfg == nil {
		return fmt.Errorf("server config not initialized")
	}

	// 获取对应名称的配置节点
	cfgNode := s.getCfg(name)
	if cfgNode.Kind == yaml.ScalarNode && cfgNode.Value == "" {
		return fmt.Errorf("config node '%s' not found", name)
	}

	// 反序列化配置到对象
	if err := cfgNode.Decode(obj); err != nil {
		return fmt.Errorf("failed to decode config '%s': %w", name, err)
	}

	return nil
}

// Run 运营一个服务器
func (s *Server) Run() error {
	if s.Basic.Version == "datetime" {
		s.Basic.Version = fmt.Sprintf("%d", time.Now().Unix())
	}
	// s.Renders = make(map[string]render.Render)
	if s.Renders["html"] == nil {
		s.Renders["html"] = new(render.HtmlRender)
		var tempFuncMap = make(template.FuncMap)
		for _, bc := range s.initCtrl {
			bc.Init(s)
		}
		for _, bc := range s.initCtrlNew {
			err := bc.Init(s)
			if err != nil {
				return err
			}
		}
		for _, v := range s.funcs {
			tempFunc := func() int {
				return 0
			}
			tempFuncMap[v.Name] = tempFunc
		}
		s.Renders["html"].Init(s, tempFuncMap)
	}
	if s.Renders["json"] == nil {
		s.Renders["json"] = new(render.JsonRender)
	}
	if s.Renders["raw"] == nil {
		s.Renders["raw"] = new(render.RawRender)
	}
	if s.Renders["xml"] == nil {
		s.Renders["xml"] = new(render.XmlRender)
	}
	slog.Info("Server starting", "port", s.Basic.Port)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Basic.Port),
		Handler:      s,
		WriteTimeout: time.Second * time.Duration(s.Basic.WriteT0),
		ReadTimeout:  time.Second * time.Duration(s.Basic.ReadT0),
	}
	srv.SetKeepAlivesEnabled(s.Basic.EnableKeepAlive)
	var err error
	if s.Basic.HttpsEnable {
		err = srv.ListenAndServeTLS(s.Basic.HttpsCertFile, s.Basic.HttpsKey)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil {
		slog.Error("Server error", "error", err)
	}
	return err
}
