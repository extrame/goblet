package goblet

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log/slog"
	"net"
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

	"github.com/extrame/goblet/v2/config"
	ge "github.com/extrame/goblet/v2/error"
	"github.com/extrame/goblet/v2/render"
	yaml "gopkg.in/yaml.v3"
)

var NotImplemented = fmt.Errorf("this method is not implemented")

type Fn struct {
	Name string
	Fn   interface{}
}

type controllerNeedInit interface {
	Init(*Server)
}

type controllerNeedInitAndReturnError interface {
	Init(*Server) error
}

// Server 服务器类型
type Server struct {
	ConfigFile string

	Config  config.Config
	router  router
	Renders map[string]render.Render

	Name          string
	plugins       map[string]NewPlugin
	funcs         []Fn
	inits         []func(*Server)
	initsNew      []func(*Server) error
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
	DB            *gorm.DB
	ctx           context.Context
	cancel        context.CancelFunc
}

func (s *Server) SetDefaultOk(fn func(*Context)) {
	s.okFunc = fn
}

func (s *Server) SetDefaultError(fn func(*Context, error, ...string)) {
	s.errFunc = fn
}

func (s *Server) Context() context.Context {
	return s.ctx
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

	var wrapper = &jsonRenderWrapper{
		successCode: 200,
		successMsg:  "success",
	}

	plugins = append(plugins, wrapper)

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
		if rv, ok := plugin.(render.Render); ok {
			s.Renders[rv.Type()] = rv
		}
		if rv, ok := plugin.(JsonRenderCodeSetter); ok {
			wrapper.successCode = rv.RespondJsonWithSuccessCode()
		}
		if rv, ok := plugin.(JsonRenderSuccessMsgSetter); ok {
			wrapper.successMsg = rv.RespondJsonWithSuccessMsg()
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
			s.Config.Db.Pwd = dbPwdPlugin.SetPwd(s.Config.Db.Pwd)
		}
		if dbUserPlugin != nil {
			s.Config.Db.User = dbUserPlugin.SetName(s.Config.Db.User)
		}
		db, err := s.connectDB()
		if err == nil {
			s.DB = db
			if s.Config.Basic.Env == config.DevelopEnv {
				db.Debug()
			}
		} else if err != config.NoDbDriver {
			slog.Error("connect error", "err", err)
		}
	} else {
		slog.With(err).Error("read config file error")
		os.Exit(1)
	}
	s.enableDbCache()
}

func (s *Server) isSilence(u string) bool {
	si, ok := s.silenceUrls[u]
	return ok && si
}

func (s *Server) connectDB() (*gorm.DB, error) {
	return s.Config.Db.New(s.Config.Basic.DbEngine)
}

// ControlBy 函数用于将控制器添加到服务器中
//
// 参数：
// controller - 需要添加到服务器中的控制器接口
//
// 函数逻辑：
// 1. 通过 wrapController 函数将控制器转换为配置信息
// 2. 如果控制器实现了 controllerNeedInit 接口，则将其添加到初始化控制器列表中
// 3. 如果控制器实现了 controllerNeedInitAndReturnError 接口，则将其添加到新的初始化控制器列表中
// 4. 将配置信息添加到路由表中
func (s *Server) ControlBy(controllers ...interface{}) {
	for _, controller := range controllers {
		cfg := s.wrapController(controller)

		if bc, ok := controller.(controllerNeedInit); ok {
			s.inits = append(s.inits, bc.Init)
		}
		if bc, ok := controller.(controllerNeedInitAndReturnError); ok {
			s.initsNew = append(s.initsNew, bc.Init)
		}
		s.router.add(cfg)
	}
}

func (s *Server) Setup(extraInit ...interface{}) error {
	if len(extraInit) > 0 {
		for _, initFn := range extraInit {
			if fn, ok := initFn.(func(*Server) error); ok {
				s.initsNew = append(s.initsNew, fn)
			} else if fn, ok := initFn.(func(*Server)); ok {
				s.inits = append(s.inits, fn)
			}
		}
	}
	return nil
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
	var sync = true

	if len(syncs) > 0 {
		sync = syncs[0]
	}

	if sync && s.Config.Basic.DbEngine != "none" {
		err := s.DB.AutoMigrate(models)
		if err != nil {
			slog.Error("migrate error:", "error", err)
			os.Exit(1)
		}
	}
}

func (s *Server) Env() string {
	return s.Config.Basic.Env
}

// Debug 当服务器环境为调试环境时，执行相应的匿名函数，用于编写调试环境专用的代码块
func (s *Server) Debug(fn func()) {
	if s.Env() == config.DevelopEnv {
		fn()
	}
}

func (s *Server) WwwRoot() string {
	if abs, err := filepath.Abs(s.Config.Basic.WwwRoot); err == nil {
		return abs
	}
	return s.Config.Basic.WwwRoot
}

func (s *Server) GetServerPathByCtrl(ctrl interface{}) []string {
	root := s.WwwRoot()
	cfg := s.wrapController(ctrl)
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
			slog.Debug("use static file name return by dynamic", "method", geE.Method)
			file := filepath.Join(s.Config.Basic.WwwRoot, s.PublicDir(), geE.Method)
			if _, err := os.Stat(file); !os.IsNotExist(err) {
				s.ServeFile(w, r, filepath.Join(s.Config.Basic.WwwRoot, s.PublicDir(), geE.Method))
				return
			}
		}
		if strings.HasSuffix(r.URL.Path, "/") {
			path = r.URL.Path + "index.html"
		} else {
			path = r.URL.Path
		}
		s.ServeFile(w, r, filepath.Join(s.Config.Basic.WwwRoot, s.PublicDir(), path))
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

func (s *Server) parseConfig() (err error) {
	reader, err := s.configer.GetConfigSource(s)
	if err == nil {
		s.initLog()
		s.cfg = new(yaml.Node)
		err = yaml.NewDecoder(reader).Decode(s.cfg)
		if err == nil {
			if err = s.cfg.Decode(&s.Config); err == nil {
				if s.Config.Db.Host == "" {
					s.Config.Db.Host = s.Name
				}
			}
		}
	}

	if err != nil {
		return err
	}

	if s.Config.Basic.Env == "" {
		s.Config.Basic.Env = config.DevelopEnv
		slog.Info("environment not set, default set as development")
	}

	if s.Config.Basic.DbEngine == "" {
		s.Config.Basic.DbEngine = "none"
	}

	if s.Config.Basic.Port == 0 {
		s.Config.Basic.Port = 8080
	}

	if s.Config.Basic.Env != config.DevelopEnv && s.Config.Basic.Env != config.ProductEnv && s.Config.Basic.Env != config.OldProductEnv {
		slog.Error("environment must be development or production, config env: development or env: production")
		os.Exit(1)
	} else if s.Config.Basic.Env == config.OldProductEnv {
		s.Config.Basic.Env = config.ProductEnv
		fmt.Println("[Deprecatd]production environment must be set as 'production' instead of 'product'")
	}
	if s.Config.Basic.Env == config.DevelopEnv {
		//enable debug in slog TODO
	}
	for _, plugin := range s.plugins {
		if err = plugin.AddCfgAndInit(s); err != nil {
			slog.Error("add plugin config error", "plugin", plugin, "error", err)
		}
	}
	return
}

// Hash 获得一个字符串的加密版本
func (s *Server) Hash(str string) string {
	hash := sha1.New()
	hash.Write([]byte(str))
	hash.Write([]byte(s.Config.Basic.HashSecret))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// PublicDir 获得服务器对应的公共文件夹的地址
func (s *Server) PublicDir() string {
	return s.Config.Basic.PublicDir
}

func (s *Server) enableDbCache() {
	// GORM 内置了缓存机制，这里可以留空或根据需求实现
}

func (s *Server) GetDelims() []string {
	return s.delims
}

// Run 运营一个服务器
func (s *Server) Run(ctx ...context.Context) error {
	if len(ctx) > 0 {
		s.ctx = ctx[0]
	} else {
		s.ctx = context.Background()
	}
	if s.Config.Basic.Version == "datetime" {
		s.Config.Basic.Version = fmt.Sprintf("%d", time.Now().Unix())
	}
	if s.Renders["json"] == nil {
		s.Renders["json"] = new(render.JsonRender)
	}
	if s.Renders["raw"] == nil {
		s.Renders["raw"] = new(render.RawRender)
	}
	for _, bc := range s.inits {
		bc(s)
	}
	for _, bc := range s.initsNew {
		if err := bc(s); err != nil {
			return err
		}
	}
	slog.With("port", s.Config.Basic.Port).Info("Listening")
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Config.Basic.Port),
		Handler:      s,
		WriteTimeout: time.Second * time.Duration(s.Config.Basic.WriteT0),
		ReadTimeout:  time.Second * time.Duration(s.Config.Basic.ReadT0),
		BaseContext: func(netListener net.Listener) context.Context {
			return s.ctx
		},
	}
	srv.SetKeepAlivesEnabled(s.Config.Basic.EnableKeepAlive)
	var err error
	if s.Config.Basic.HttpsEnable {
		err = srv.ListenAndServeTLS(s.Config.Basic.HttpsCertFile, s.Config.Basic.HttpsKey)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		slog.Error(err.Error())
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func (s *Server) GetDefaultRender() string {
	return s.defaultRender
}
