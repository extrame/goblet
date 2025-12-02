package goblet

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// EnableSse 启用SSE连接，设置必要的响应头
func (c *Context) EnableSse() error {
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

	return nil
}

// SseSend 发送SSE消息
// message: 要发送的消息内容，可以是任意类型，会被转换为字符串
// action: 可选的事件类型，如果提供则作为event字段发送
func (c *Context) SseSend(message interface{}, action ...string) error {
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
	_, err := c.writer.Write([]byte(sseMessage))
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
func (c *Context) SseEnd(message ...string) error {
	// 构建结束消息，使用标准的[DONE]信号
	endMessage := "[DONE]"
	if len(message) > 0 && message[0] != "" {
		endMessage = fmt.Sprintf("[DONE] %s", message[0])
	}

	// 发送结束事件
	sseMessage := fmt.Sprintf("data: %s\n\n", endMessage)

	// 写入响应
	_, err := c.writer.Write([]byte(sseMessage))
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
