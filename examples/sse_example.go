package main

import (
	"fmt"
	"time"

	"github.com/extrame/goblet/v2"
)

// TestedController 使用Goblet控制器模式实现SSE功能
type TestedController struct {
	goblet.SingleController `Route:"/sse"`
}

// GetSimple 简单的SSE端点示例
// Route: GET /sse/simple
func (c *TestedController) Simple(ctx *goblet.Context) {
	// 启用SSE连接
	if err := ctx.EnableSse(); err != nil {
		ctx.RespondError(err, "SSE启用失败")
		return
	}

	// 发送欢迎消息
	if err := ctx.SseSend("欢迎使用SSE！", "welcome"); err != nil {
		fmt.Printf("发送消息失败: %v\n", err)
		return
	}

	// 模拟实时数据推送
	for i := 0; i < 10; i++ {
		message := fmt.Sprintf("当前计数: %d", i)
		if err := ctx.SseSend(message, "count"); err != nil {
			fmt.Printf("发送消息失败: %v\n", err)
			return
		}
		time.Sleep(1 * time.Second)
	}

	// 发送结束消息
	ctx.SseSend("数据推送完成！", "complete")

	// 使用SseEnd优雅地结束连接，发送标准的[DONE]信号
	time.Sleep(500 * time.Millisecond) // 稍微等待确保消息发送完成
	ctx.SseEnd("SSE连接正常结束")
}

// GetJson 使用JSON格式的SSE端点示例
// Route: GET /sse/json
func (c *TestedController) GetJson(ctx *goblet.Context) {
	if err := ctx.EnableSse(); err != nil {
		ctx.RespondError(err, "SSE启用失败")
		return
	}

	// 发送JSON数据
	data := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"message":   "实时数据更新",
		"status":    "running",
	}

	if err := ctx.SseSendJSON(data, "update"); err != nil {
		fmt.Printf("发送JSON消息失败: %v\n", err)
		return
	}
}

// GetError 错误处理示例
// Route: GET /sse/error
func (c *TestedController) GetError(ctx *goblet.Context) {
	if err := ctx.EnableSse(); err != nil {
		ctx.RespondError(err, "SSE启用失败")
		return
	}

	// 模拟错误情况
	if err := fmt.Errorf("模拟错误"); err != nil {
		if sendErr := ctx.SseSendError(err, "system_error"); sendErr != nil {
			fmt.Printf("发送错误消息失败: %v\n", sendErr)
		}
		return
	}
}

// GetMultiline 多行消息示例
// Route: GET /sse/multiline
func (c *TestedController) Multiline(ctx *goblet.Context) {
	if err := ctx.EnableSse(); err != nil {
		ctx.RespondError(err, "SSE启用失败")
		return
	}

	// 发送多行消息
	multilineMessage := `这是第一行
这是第二行
这是第三行`

	if err := ctx.SseSend(multilineMessage, "multiline"); err != nil {
		fmt.Printf("发送多行消息失败: %v\n", err)
		return
	}
}

// SSE使用示例
func main() {
	// 创建Goblet服务器实例
	server := goblet.Organize("sse_example")

	// 注册TestedController，自动注册所有路由
	server.ControlBy(&TestedController{})

	// 启动服务器
	server.Run()
}

/*
使用示例：

1. 测试简单SSE连接：
curl -N http://localhost:8080/sse/simple

2. 测试JSON格式的SSE：
curl -N http://localhost:8080/sse/json

3. 测试错误处理：
curl -N http://localhost:8080/sse/error

4. 测试多行消息：
curl -N http://localhost:8080/sse/multiline

5. HTML页面中使用SSE：
```html
<!DOCTYPE html>
<html>
<head>
    <title>SSE测试</title>
</head>
<body>
    <div id="messages"></div>
    <script>
        const eventSource = new EventSource('/sse/simple');

        eventSource.onmessage = function(event) {
            document.getElementById('messages').innerHTML += '<p>' + event.data + '</p>';
        };

        eventSource.addEventListener('welcome', function(event) {
            console.log('收到欢迎消息:', event.data);
        });

        eventSource.addEventListener('count', function(event) {
            console.log('收到计数消息:', event.data);
        });

        eventSource.onerror = function(error) {
            console.error('SSE错误:', error);
        };
    </script>
</body>
</html>
```

控制器模式说明：
- 使用 TestedController 结构体嵌入 goblet.SingleController
- 控制器路由基路径为 "/sse"（通过 Route 标签指定）
- 方法名格式：HTTP方法 + 路径片段
  - GetSimple -> GET /sse/simple
  - GetJson -> GET /sse/json
  - GetError -> GET /sse/error
  - GetMultiline -> GET /sse/multiline
*/
