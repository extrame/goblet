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
```

---

## RestController 模式

### 概述

`goblet.RestController` 提供了一套完整的 RESTful API 模板，自动为资源生成 7 个标准端点，并支持灵活的方法签名和扩展子路径。

### 快速开始

```go
type TaskController struct {
    goblet.RestController `Route:"/tasks"`
}
```

### 标准RESTful端点

注册后自动映射以下7个标准端点：

| HTTP方法  | 路径              | 控制器方法      | 说明          |
|-----------|-------------------|----------------|--------------|
| GET       | /tasks            | Index()        | 获取资源列表    |
| GET       | /tasks/new        | New()          | 新建表单页     |
| POST      | /tasks            | Create()       | 创建资源       |
| GET       | /tasks/:id        | Show()         | 获取单个资源    |
| GET       | /tasks/:id/edit   | Edit()         | 编辑表单页     |
| PATCH/PUT | /tasks/:id        | Update()       | 更新资源       |
| DELETE    | /tasks/:id        | Destroy()      | 删除资源       |

### 灵活的方法签名

方法参数按以下顺序自动注入：

1. **路径参数** — `string`/`int`/`uint` 类型，从URL路径自动提取
2. **`*goblet.Context`** — 上下文对象
3. **请求体结构体** — 从请求体 JSON 自动解析（`*struct` 或 `struct`）
4. **`map` 类型** — 从请求体 JSON 自动解析

支持的方法签名示例：

```go
// 只有Context
func (c *TaskController) Index(ctx *goblet.Context) ([]Task, error)

// 路径参数 + Context
func (c *TaskController) Show(id string, ctx *goblet.Context) (*Task, error)

// Context + 请求体
func (c *TaskController) Create(ctx *goblet.Context, req *CreateRequest) (*Task, error)

// 路径参数 + Context + 请求体
func (c *TaskController) Update(id string, ctx *goblet.Context, req *UpdateRequest) (*Task, error)

// 仅返回error
func (c *TaskController) Destroy(id string, ctx *goblet.Context) error

// 无返回值，通过ctx手动设置响应
func (c *TaskController) New(ctx *goblet.Context) error {
    return nil
}
```

返回值支持以下组合：

- `(data, error)` — 数据和错误
- `(data)` — 仅数据
- `(error)` — 仅错误
- 无返回值 — 通过 `ctx.Respond()` 手动设置

### 扩展子路径方法（新特性）

在标准RESTful端点基础上，支持按子路径名称扩展方法。

#### GET 路径扩展

当访问 `GET /tasks/done` 时，Rest控制器会：

1. **优先尝试** 查找 `IndexDone()` 方法（`Index` + `Done` 的组合）
2. **若不存在**，回退到标准 `Show("done", ctx)` 方法

```go
// GET /tasks/done -> 调用此方法
func (c *TaskController) IndexDone(ctx *goblet.Context) ([]Task, error) {
    // ctx.DB.Where("status = ?", "done").Find(&tasks)
    return tasks, nil
}

// GET /tasks/pending -> 调用此方法
func (c *TaskController) IndexPending(ctx *goblet.Context) ([]Task, error) {
    // ctx.DB.Where("status = ?", "pending").Find(&tasks)
    return tasks, nil
}
```

#### POST 路径扩展

当访问 `POST /tasks/123/approve` 时，Rest控制器：

1. 提取 id = `"123"`，剩余路径 `/approve`
2. 调用 `CreateApprove(id, ctx, req)` 方法

```go
// POST /tasks/:id/approve
func (c *TaskController) CreateApprove(id string, ctx *goblet.Context, req *CreateTaskRequest) (*Task, error) {
    var task Task
    // ctx.DB.First(&task, id)
    // task.Status = "done"
    // ctx.DB.Save(&task)
    return &task, nil
}

// POST /tasks/:id/reject
func (c *TaskController) CreateReject(id string, ctx *goblet.Context, req *CreateTaskRequest) (*Task, error) {
    var task Task
    // ctx.DB.First(&task, id)
    // task.Status = "rejected"
    // ctx.DB.Save(&task)
    return &task, nil
}
```

### 完整示例

完整的可运行示例请参考：

- [Rest示例代码](../examples/rest/main.go)
- [SSE示例代码](../examples/sse_example.go)
- [基础示例代码](../examples/example/main.go)
