# 控制器写法

## 定义结构体

```go
type MyController struct {
    goblet.Route `path:"/mycontroller"`
}
```

其中，`goblet.Route` 是一个结构体标签，用于定义路由信息。

## 实现方法

```go
func (c *MyController) Get(id int, ctx *goblet.Context) error {
    return nil
}
```

这里，`Get` 方法是一个 HTTP GET 请求的处理函数。
ctx 参数包含了请求的上下文信息，例如路径参数、查询字符串等。
在ctx之前可以定义多个参数，他们将在路由参数中被解析。
例如： /mycontroller/12345
id 将被解析为12345
返回可以是一个错误，也可以是任何你想要返回的内容加错误，下面的定义方式都是可以的
```go
func (c *MyController) Get(id int, ctx *goblet.Context) error {
    return nil
}
```

```go
func (c *MyController) Get(id int, ctx *goblet.Context) (int, error) {
    return 1, nil
}
```

返回的错误将被传递给 Goblet 的错误处理机制。在默认的JSON渲染器中，它将返回一个包含错误信息的JSON响应。
```json
{
    "msg": "some error message",
    "code": 400
}
```
如果错误为空，它将使用第一个返回内容作为data字段
```json
{
    "data": 1,
    "code": 200,
    "msg": "ok"
}
```

## 鉴权AOP

控制器可以定义Init函数，并且在其中调用server.Pre方法，将任何前置函数注册在控制函数之前执行，前置函数可以定义登录检查等逻辑
```go
func (c *Controller) Init(server *goblet.Server) {
	server.Pre(checkLogin, "Func1")
}

func checkLogin(ctx *goblet.Context) error {
	user, ok := ctx.GetLoginInfo()
	if !ok {
		return goblet.Interrupted("no login user")
	}
	ctx.AddInfo("user_id", user.Id)
	return nil
}
```

### 定义特殊错误类型：

1. goblet.Interrupted(reason)：该错误类型会中断请求处理，直接返回错误信息给客户端，不会执行后续的业务逻辑。


## 路径匹配

### 路径参数

在路由定义中，可以使用 `:param` 的形式来捕获URL中的动态部分作为参数。例如：

```go
type MyController struct {
    goblet.Route `path:"/mycontroller/:id"`
}
