package ctrl

import (
	"time"

	"github.com/extrame/goblet/v2"
)

// ============================================================
// 模型定义 / Model Definition
// ============================================================

// Task 任务模型
type Task struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title" binding:"required"`
	Status    string    `json:"status"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================================
// 请求/响应结构体定义
// ============================================================

type CreateTaskRequest struct {
	Title    string `json:"title" binding:"required"`
	Priority int    `json:"priority"`
}

type UpdateTaskRequest struct {
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority int    `json:"priority"`
}

// ============================================================
// RestController 模式示例 / Rest Mode Example
// ============================================================

// TaskController 任务控制器
// Rest模式自动生成7个标准RESTful端点:
//
//	Index   -> GET    /tasks
//	Create  -> POST   /tasks
//	Show    -> GET    /tasks/:id
//	Update  -> PATCH  /tasks/:id
//	Destroy -> DELETE /tasks/:id
//	New     -> GET    /tasks/new
//	Edit    -> GET    /tasks/:id/edit
//
// 【新特性】支持扩展子路径方法，在标准REST基础上提高灵活性
//
//	GET  /tasks/done      -> IndexDone()
//	POST /tasks/:id/approve -> CreateApprove(id, ctx, req)
type TaskController struct {
	goblet.RestController `Route:"/tasks"`
}

// ---- 标准RESTful方法 ----

// Index 获取任务列表
// Route: GET /tasks
func (c *TaskController) Index(ctx *goblet.Context) ([]Task, error) {
	var tasks []Task
	// ctx.DB.Find(&tasks)
	return tasks, nil
}

// Show 获取单个任务
// Route: GET /tasks/:id
// id 自动从URL路径中提取
func (c *TaskController) Show(id string, ctx *goblet.Context) (*Task, error) {
	var task Task
	// ctx.DB.First(&task, id)
	return &task, nil
}

// Create 创建任务
// Route: POST /tasks
// req 从请求体JSON自动解析
func (c *TaskController) Create(ctx *goblet.Context, req *CreateTaskRequest) (*Task, error) {
	task := Task{
		Title:    req.Title,
		Status:   "pending",
		Priority: req.Priority,
	}
	// ctx.DB.Create(&task)
	return &task, nil
}

// Update 更新任务
// Route: PATCH/PUT /tasks/:id
func (c *TaskController) Update(id string, ctx *goblet.Context, req *UpdateTaskRequest) (*Task, error) {
	var task Task
	// ctx.DB.First(&task, id); ctx.DB.Model(&task).Updates(req)
	return &task, nil
}

// Destroy 删除任务
// Route: DELETE /tasks/:id
func (c *TaskController) Destroy(id string, ctx *goblet.Context) error {
	// ctx.DB.Delete(&Task{}, id)
	return nil
}

// New 新建任务表单页
// Route: GET /tasks/new
func (c *TaskController) New(ctx *goblet.Context) error {
	return nil
}

// Edit 编辑任务表单页
// Route: GET /tasks/:id/edit
func (c *TaskController) Edit(id string, ctx *goblet.Context) error {
	return nil
}

// ---- 扩展子路径方法（新特性）----

// IndexDone 获取已完成任务列表
// Route: GET /tasks/done
// 说明：/tasks/done 匹配标准REST的 /tasks/:id，
// Rest控制器会优先尝试查找 IndexDone() 方法，
// 存在则调用，不存在则回退到 Show("done", ctx)
func (c *TaskController) IndexDone(ctx *goblet.Context) ([]Task, error) {
	var tasks []Task
	// ctx.DB.Where("status = ?", "done").Find(&tasks)
	return tasks, nil
}

// IndexPending 获取待办任务列表
// Route: GET /tasks/pending
func (c *TaskController) IndexPending(ctx *goblet.Context) ([]Task, error) {
	var tasks []Task
	// ctx.DB.Where("status = ?", "pending").Find(&tasks)
	return tasks, nil
}

// CreateApprove 审批任务（通过ID）
// Route: POST /tasks/:id/approve
// 【新特性】灵活的方法签名：
// - id: URL路径参数，自动提取
// - ctx: Context上下文
// - req: 请求体，自动JSON解析
func (c *TaskController) CreateApprove(id string, ctx *goblet.Context, req *CreateTaskRequest) (*Task, error) {
	var task Task
	// ctx.DB.First(&task, id); task.Status = "done"; ctx.DB.Save(&task)
	return &task, nil
}

// CreateReject 驳回任务（通过ID）
// Route: POST /tasks/:id/reject
func (c *TaskController) CreateReject(id string, ctx *goblet.Context, req *CreateTaskRequest) (*Task, error) {
	var task Task
	// ctx.DB.First(&task, id); task.Status = "rejected"; ctx.DB.Save(&task)
	return &task, nil
}

// ============================================================
// 灵活的方法签名说明 / Method Signature Flexibility
// ============================================================
// Rest模式支持多种方法签名，参数按以下顺序自动注入：
//
// 1. 路径参数: string/int/uint类型，从URL路径自动提取
// 2. Context参数
// 3. 请求体*结构体: 从JSON自动解析
//
// 示例:
//
//	func Index(ctx) (data, error)
//	func Show(id string, ctx) (data, error)
//	func Create(ctx, req *CreateRequest) (data, error)
//	func Update(id string, ctx, req *UpdateRequest) (data, error)
//	func CreateApprove(id string, ctx, req *CreateTaskRequest) (data, error)

// ============================================================

func main() {
	server := goblet.Organize("rest_example")
	server.ControlBy(&TaskController{})
	server.Run()
}

/*
# Rest模式路由对照表

## 标准RESTful端点

| HTTP方法  | 路径              | 控制器方法      |
|-----------|-------------------|----------------|
| GET       | /tasks            | Index()        |
| GET       | /tasks/new        | New()          |
| POST      | /tasks            | Create()       |
| GET       | /tasks/:id        | Show()         |
| GET       | /tasks/:id/edit   | Edit()         |
| PATCH/PUT | /tasks/:id        | Update()       |
| DELETE    | /tasks/:id        | Destroy()      |

## 扩展子路径端点（新特性）

| HTTP方法 | 路径                | 控制器方法           |
|----------|-------------------|---------------------|
| GET      | /tasks/done       | IndexDone()         |
| GET      | /tasks/pending    | IndexPending()      |
| POST     | /tasks/:id/approve | CreateApprove()     |
| POST     | /tasks/:id/reject  | CreateReject()      |

## curl测试示例

# 获取所有任务
curl http://localhost:8080/tasks

# 获取单个任务
curl http://localhost:8080/tasks/1

# 创建任务
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"完成任务","priority":1}'

# 更新任务
curl -X PATCH http://localhost:8080/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"status":"done"}'

# 删除任务
curl -X DELETE http://localhost:8080/tasks/1

# 获取已完成任务（扩展子路径）
curl http://localhost:8080/tasks/done

# 审批任务（扩展子路径 + 请求体）
curl -X POST http://localhost:8080/tasks/1/approve \
  -H "Content-Type: application/json" \
  -d '{"title":"审批通过","priority":1}'
*/
