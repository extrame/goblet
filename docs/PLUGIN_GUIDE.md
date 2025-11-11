# Goblet 插件开发指南

## 插件接口

Goblet 框架支持多种类型的插件，以下是主要的插件接口：

### 1. 基础插件接口
```go
type NewPlugin interface {
    AddCfgAndInit(*Server) error
}
```

### 2. 特殊功能插件接口
```go
// 数据库密码插件
type DbPwdPlugin interface {
    SetPwd(string) string
}

// 数据库用户名插件
type dbUserNamePlugin interface {
    SetName(string) string
}

// 请求处理插件
type onNewRequestPlugin interface{}

// 存储插件
type Saver interface{}

// KV存储驱动
type KvDriver interface{}

// 成功响应处理器
type OkFuncSetter interface {
    RespendOk(*Context)
}

// 错误响应处理器
type ErrFuncSetter interface {
    RespondError(*Context, error, ...string)
}

// 默认渲染器设置
type DefaultRenderSetter interface {
    DefaultRender() string
}

// 静默URL设置
type SilenceUrlSetter interface {
    SetSilenceUrls() map[string]bool
}

// 登录信息存储
type LoginInfoStorer interface{}

// 配置器
type Configer interface {
    GetConfigSource(*Server) (io.Reader, error)
}

// 分隔符设置
type DelimSetter interface {
    SetDelim() []string
}

// 渲染器
type Render interface {
    Type() string
    PrepareInstance(*Context) (RenderInstance, error)
}
```

## 插件开发步骤

### 1. 定义插件结构体
```go
type MyPlugin struct {
    // 插件配置字段
}
```

### 2. 实现所需接口
```go
func (p *MyPlugin) AddCfgAndInit(s *Server) error {
    // 初始化逻辑
    return nil
}
```

### 3. 注册插件
在Server初始化时通过Organize方法注册插件：
```go
server.Organize("myapp", []interface{}{&MyPlugin{}})
```

## 插件使用

### 获取插件实例
```go
plugin := server.GetPlugin("myplugin")
if plugin != nil {
    // 使用插件
}
```

## 最佳实践

1. 插件命名应清晰明确，使用小写
2. 复杂的插件应该拆分为多个小插件
3. 插件应该处理自己的配置和错误
4. 避免插件之间的直接依赖

## 示例插件

```go
// 日志插件示例
type LogPlugin struct {
    Level string `yaml:"level"`
}

func (p *LogPlugin) AddCfgAndInit(s *Server) error {
    level, err := slog.ParseLevel(p.Level)
    if err != nil {
        return err
    }
    slog.SetLogLoggerLevel(level)
    return nil
}

// 注册插件
server.Organize("myapp", []interface{}{&LogPlugin{Level: "debug"}})
```