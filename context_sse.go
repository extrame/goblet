package goblet

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type SseOption func(*Context)

// sseKeepAliveTimer 存储SSE保活定时器信息
type sseKeepAliveTimer struct {
	timer    *time.Timer
	timeout  time.Duration
	mu       sync.Mutex
	started  bool
	stopChan chan bool
}

func SseWithKeepAlive(timeout int) SseOption {
	return func(c *Context) {
		if timeout <= 0 {
			return
		}

		// 创建保活定时器信息
		keepAlive := &sseKeepAliveTimer{
			timeout:  time.Duration(timeout) * time.Second,
			stopChan: make(chan bool, 1),
		}

		// 存储在context的extra中
		if c.extra == nil {
			c.extra = make(map[string]interface{})
		}
		c.extra["sse_keepalive"] = keepAlive
	}
}

// stopKeepAliveTimer 停止KeepAlive定时器
func (c *Context) stopKeepAliveTimer() {
	if c.extra == nil {
		return
	}

	keepAliveInterface, exists := c.extra["sse_keepalive"]
	if !exists {
		return
	}

	keepAlive, ok := keepAliveInterface.(*sseKeepAliveTimer)
	if !ok {
		return
	}

	keepAlive.mu.Lock()
	defer keepAlive.mu.Unlock()

	// 停止定时器
	if keepAlive.timer != nil {
		keepAlive.timer.Stop()
		keepAlive.timer = nil
	}

	// 标记为已停止
	keepAlive.started = false
}

// EnableSse 启用SSE连接，设置必要的响应头
func (c *Context) EnableSse(options ...SseOption) error {
	// 设置SSE所需的响应头
	c.writer.Header().Set("Content-Type", "text/event-stream")
	c.writer.Header().Set("Cache-Control", "no-cache")
	c.writer.Header().Set("Connection", "keep-alive")
	c.writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 写入状态码
	c.writer.WriteHeader(http.StatusOK)

	// 刷新响应，确保头部立即发送
	if flusher, ok := c.writer.(http.Flusher); ok {
		flusher.Flush()
	} else {
		return fmt.Errorf("response writer does not support flushing")
	}

	// 应用SSE选项
	for _, option := range options {
		option(c)
	}

	return nil
}

// startKeepAliveTimer 启动或重置KeepAlive定时器
func (c *Context) startKeepAliveTimer() {
	if c.extra == nil {
		return
	}

	keepAliveInterface, exists := c.extra["sse_keepalive"]
	if !exists {
		return
	}

	keepAlive, ok := keepAliveInterface.(*sseKeepAliveTimer)
	if !ok {
		return
	}

	keepAlive.mu.Lock()
	defer keepAlive.mu.Unlock()

	// 如果已经存在定时器，先停止它
	if keepAlive.timer != nil {
		keepAlive.timer.Stop()
	}

	// 创建新的定时器
	keepAlive.timer = time.AfterFunc(keepAlive.timeout, func() {
		// 防止panic导致程序崩溃
		defer func() {
			if r := recover(); r != nil {
				// 记录错误但不影响其他goroutine
				slog.Error("panic in SSE keepalive timer", "error", r)
			}
		}()

		// 发送KeepAlive消息
		if err := c.SseSend(":keepalive\n\n"); err != nil {
			// 如果发送失败，可以选择记录日志或停止定时器
			slog.Error("failed to send keepalive message", "error", err)
			return
		}

		// 重新启动定时器
		keepAlive.mu.Lock()
		if keepAlive.started {
			keepAlive.timer.Reset(keepAlive.timeout)
		}
		keepAlive.mu.Unlock()
	})

	// 标记为已启动
	if !keepAlive.started {
		keepAlive.started = true
	}
}

// SseSend 发送SSE消息
// message: 要发送的消息内容，可以是任意类型，会被转换为字符串
// action: 可选的事件类型，如果提供则作为event字段发送
func (c *Context) SseSend(message interface{}, action ...string) (err error) {

	select {
	case <-c.Request.Context().Done():
		return fmt.Errorf("context canceled: %w", c.Request.Context().Err())
	default:
	}

	// 在发送消息后重置KeepAlive定时器
	defer c.startKeepAliveTimer()
	// 转换消息为字符串
	var messageStr string
	switch v := message.(type) {
	case string:
		messageStr = v
	case []byte:
		messageStr = string(v)
	default:
		messageStr = fmt.Sprintf("%v", v)
	}

	// 构建SSE格式的消息
	var sseMessage string

	// 添加事件类型（如果有）
	if len(action) > 0 && action[0] != "" {
		sseMessage += fmt.Sprintf("event: %s\n", action[0])
	}

	// 添加数据（支持多行消息）
	lines := splitLines(messageStr)
	for _, line := range lines {
		sseMessage += fmt.Sprintf("data: %s\n", line)
	}

	// 添加消息结束标记
	sseMessage += "\n"

	// 写入响应
	_, err = c.writer.Write([]byte(sseMessage))
	if err != nil {
		return fmt.Errorf("failed to write SSE message: %w", err)
	}

	// 刷新响应，确保消息立即发送
	if flusher, ok := c.writer.(http.Flusher); ok {
		flusher.Flush()
	} else {
		return fmt.Errorf("response writer does not support flushing")
	}

	return nil
}

// splitLines 将字符串按行分割，处理不同平台的换行符
func splitLines(s string) []string {
	var lines []string
	var currentLine []rune

	for _, r := range s {
		if r == '\n' {
			lines = append(lines, string(currentLine))
			currentLine = []rune{}
		} else if r == '\r' {
			// 跳过\r，等待\n
			continue
		} else {
			currentLine = append(currentLine, r)
		}
	}

	// 添加最后一行（如果有）
	if len(currentLine) > 0 {
		lines = append(lines, string(currentLine))
	}

	// 如果输入为空，返回空字符串的数组
	if len(lines) == 0 {
		lines = []string{""}
	}

	return lines
}

// SseSendJSON 发送JSON格式的SSE消息（便捷方法）
func (c *Context) SseSendJSON(data interface{}, action ...string) error {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return c.SseSend(jsonStr, action...)
}

// SseSendError 发送错误信息的SSE消息（便捷方法）
func (c *Context) SseSendError(err error, action ...string) error {
	errorAction := "error"
	if len(action) > 0 && action[0] != "" {
		errorAction = action[0]
	}
	return c.SseSend(fmt.Sprintf("Error: %v", err), errorAction)
}

// SseEnd 发送SSE结束信号并关闭连接
// 发送标准的"[DONE]"信号，这是SSE的通用结束约定
func (c *Context) SseEnd(message ...string) (err error) {

	// 停止KeepAlive定时器
	defer c.stopKeepAliveTimer()
	// 构建结束消息，使用标准的[DONE]信号
	endMessage := "[DONE]"
	if len(message) > 0 && message[0] != "" {
		endMessage = fmt.Sprintf("[DONE] %s", message[0])
	}

	// 发送结束事件
	sseMessage := fmt.Sprintf("data: %s\n\n", endMessage)

	// 写入响应
	_, err = c.writer.Write([]byte(sseMessage))
	if err != nil {
		return fmt.Errorf("failed to write SSE end message: %w", err)
	}

	// 刷新响应，确保消息立即发送
	if flusher, ok := c.writer.(http.Flusher); ok {
		flusher.Flush()
	} else {
		return fmt.Errorf("response writer does not support flushing")
	}

	// 关闭连接（通过关闭HTTP响应）
	// 注意：在HTTP/1.1中，关闭连接通常由服务器处理
	// 这里我们只是确保所有数据都被发送出去
	if closer, ok := c.writer.(interface{ Close() error }); ok {
		return closer.Close()
	}

	return nil
}
