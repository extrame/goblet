package ctrl

import (
	"fmt"
	"strings"
	"testing"
)

// 模拟 HttpMethodController 类型
type HttpMethodController byte

// 模拟控制器
type TestHttpController struct {
	HttpMethodController
}

func (c *TestHttpController) Get() string {
	return "Get"
}

func (c *TestHttpController) Post() string {
	return "Post"
}

func (c *TestHttpController) Delete() string {
	return "Delete"
}

func (c *TestHttpController) Put() string {
	return "Put"
}

func (c *TestHttpController) GetConfig() string {
	return "GetConfig"
}

func (c *TestHttpController) PostInfo() string {
	return "PostInfo"
}

// TestHttpMethodDocumentation 测试文档说明与实际行为的一致性
//
// 文档说明 (blockoption.go):
// GET /test -> Test.Get
// POST /test -> Test.Post
// DELETE /test -> Test.Delete
// PUT /test -> Test.Put
// GET /test/config -> Test.GetConfig
// POST /test/post/info -> Test.PostPostInfo
//
// http.go 中的核心逻辑:
// method := c.ReqMethod()  // 获取 HTTP 方法，如 "GET"
// suffix, _ := c.Suffix(true)  // 获取路径后缀，如 "/config" 或 "/post/info"
// suffixWithMethod := strings.ToLower(method) + suffix  // 拼接，如 "get/config" 或 "post/post/info"
//
// h.methods.Match(suffixWithMethod, ...) 用于匹配注册的路由
func TestHttpMethodDocumentation(t *testing.T) {
	tests := []struct {
		name           string
		urlPath        string
		httpMethod     string
		expectedMethod string
		shouldMatch    bool
		description    string
	}{
		{
			name:           "test-get",
			urlPath:        "/test",
			httpMethod:     "GET",
			expectedMethod: "Get",
			shouldMatch:    true,
			description:    "GET /test -> Test.Get",
		},
		{
			name:           "test-post",
			urlPath:        "/test",
			httpMethod:     "POST",
			expectedMethod: "Post",
			shouldMatch:    true,
			description:    "POST /test -> Test.Post",
		},
		{
			name:           "test-delete",
			urlPath:        "/test",
			httpMethod:     "DELETE",
			expectedMethod: "Delete",
			shouldMatch:    true,
			description:    "DELETE /test -> Test.Delete",
		},
		{
			name:           "test-put",
			urlPath:        "/test",
			httpMethod:     "PUT",
			expectedMethod: "Put",
			shouldMatch:    true,
			description:    "PUT /test -> Test.Put",
		},
		{
			name:           "test-get-config",
			urlPath:        "/test/config",
			httpMethod:     "GET",
			expectedMethod: "GetConfig",
			shouldMatch:    true,
			description:    "GET /test/config -> Test.GetConfig",
		},
		{
			name:           "test-post-post-info",
			urlPath:        "/test/post/info",
			httpMethod:     "POST",
			expectedMethod: "PostPostInfo",
			shouldMatch:    true,
			description:    "POST /test/post/info -> Test.PostPostInfo",
		},
	}

	fmt.Println("=== HttpMethodController 文档测试 ===")
	fmt.Println()

	for _, tt := range tests {
		fmt.Printf("测试: %s\n", tt.description)
		fmt.Printf("  URL: %s, HTTP Method: %s\n", tt.urlPath, tt.httpMethod)

		// 分析 http.go 中的 Parse 方法逻辑
		// 1. method := c.ReqMethod() 获取 HTTP 方法
		// 2. suffix, suffixWithSlash := c.Suffix(true)
		// 3. suffixWithMethod := strings.ToLower(method) + suffix
		// 4. h.methods.Match(suffixWithMethod, ...)

		// 模拟 c.Suffix(true) 的行为
		// suffix 格式为 "/config" 或 "/post/info"（有前导 /，从控制器路由之后的部分）
		// 例如 "/test/config" -> suffix 为 "/config"
		pathParts := strings.SplitN(tt.urlPath, "/", 3)
		var suffix string
		if len(pathParts) >= 3 {
			suffix = "/" + pathParts[2]
		}

		// http.go 中的关键逻辑
		suffixWithMethod := strings.ToLower(tt.httpMethod) + suffix

		fmt.Printf("  suffix = %q\n", suffix)
		fmt.Printf("  suffixWithMethod (h.methods.Match 匹配字符串) = %q\n", suffixWithMethod)

		// 验证：suffixWithMethod 应该能匹配到控制器方法名
		// 例如 "get/get" 应该能匹配到 "Get" 方法
		methodName := getMethodNameFromPath(tt.urlPath, tt.httpMethod)
		fmt.Printf("  期望调用的控制器方法: %s.%s()\n", tt.httpMethod, methodName)

		// 检查 suffixWithMethod 格式是否正确
		// 例如 "get/get" 需要转换为 "Get"（首字母大写）
		actualMethodName := suffixToMethodName(suffixWithMethod)
		fmt.Printf("  从 suffixWithMethod 提取的方法名: %s\n", actualMethodName)

		if strings.EqualFold(actualMethodName, tt.expectedMethod) {
			fmt.Printf("  ✓ 匹配正确\n")
		} else {
			fmt.Printf("  ✗ 不匹配: 期望 %s, 实际 %s\n", tt.expectedMethod, actualMethodName)
		}
		fmt.Println()
	}
}

// TestHttpMethodLogicAnalysis 分析 http.go Parse 方法的核心逻辑
func TestHttpMethodLogicAnalysis(t *testing.T) {
	fmt.Println("=== HttpMethod Parse 方法逻辑分析 ===")
	fmt.Println()

	// 模拟 http.go 中的 Parse 方法核心逻辑
	analysis := []struct {
		urlPath    string
		httpMethod string
	}{
		{"/test", "GET"},
		{"/test", "POST"},
		{"/test", "DELETE"},
		{"/test", "PUT"},
		{"/test/config", "GET"},
		{"/test/post/info", "POST"},
	}

	for _, a := range analysis {
		fmt.Printf("输入: URL=%s, HTTP=%s\n", a.urlPath, a.httpMethod)

		// 模拟 http.go 第27行: suffix, _ := c.Suffix(true)
		// 模拟 http.go 第30行: suffixWithMethod = strings.ToLower(method) + suffix
		pathParts := strings.SplitN(a.urlPath, "/", 3)
		var suffix string
		if len(pathParts) >= 3 {
			suffix = "/" + pathParts[2]
		} else {
			suffix = "/"
		}
		suffixWithMethod := strings.ToLower(a.httpMethod) + suffix

		fmt.Printf("  suffix = %q\n", suffix)
		fmt.Printf("  suffixWithMethod (匹配字符串) = %q\n", suffixWithMethod)

		// http.go 会调用 h.methods.Match(suffixWithMethod, len(suffixWithMethod))
		// Match 会尝试匹配已注册的方法，如 "get" -> "Get", "get/get" -> "GetGet"

		// 文档说明期望的方法名
		docExpected := getMethodNameFromPath(a.urlPath, a.httpMethod)
		fmt.Printf("  文档期望方法: %s.%s()\n", a.httpMethod, docExpected)

		// 从 suffixWithMethod 提取方法名（首字母大写）
		actual := suffixToMethodName(suffixWithMethod)
		fmt.Printf("  实际查找方法: %s.%s()\n", a.httpMethod, actual)

		if strings.EqualFold(actual, docExpected) {
			fmt.Printf("  ✓ 一致\n")
		} else {
			fmt.Printf("  ✗ 不一致！\n")
		}
		fmt.Println()
	}
}

// getMethodNameFromPath 根据 URL 路径和 HTTP 方法获取期望的控制器方法名
// 文档说明: GET /test -> Test.Get
// GET /test/config -> Test.GetConfig
// 即从 URL 路径中提取 controller 路由之后的部分（suffix）
// 然后与 HTTP 方法名拼接
func getMethodNameFromPath(urlPath, httpMethod string) string {
	// /test -> suffix = "/" -> 方法 = "Get"
	// /test/config -> suffix = "/config" -> 方法 = "Get" + "Config" = "GetConfig"
	// /test/post/info -> suffix = "/post/info" -> 方法 = "Post" + "Post" + "Info" = "PostPostInfo"
	pathParts := strings.SplitN(urlPath, "/", 3)
	var suffix string
	if len(pathParts) >= 3 {
		suffix = "/" + pathParts[2]
	} else {
		suffix = "/"
	}

	// suffixToMethodName 会将 "/config" 转换为 "Config"
	methodPart := suffixToMethodName(suffix)
	// 组合: HTTP方法 + suffix转换
	return httpMethod + methodPart
}

// suffixToMethodName 将 suffix 转换为方法名部分
// 例如 "/config" -> "Config", "/post/info" -> "PostInfo"
// 规则：去掉前导 "/"，将每一段的第一个字符大写，其余小写
func suffixToMethodName(suffix string) string {
	// 去掉前导 "/"
	if len(suffix) > 0 && suffix[0] == '/' {
		suffix = suffix[1:]
	}
	if suffix == "" {
		return ""
	}

	parts := strings.Split(suffix, "/")
	var result []string
	for _, part := range parts {
		if part != "" {
			// 每段首字母大写，其余小写
			result = append(result, strings.Title(strings.ToLower(part)))
		}
	}
	return strings.Join(result, "")
}

// TestSuffixToMethodName 测试 suffixToMethodName 函数
func TestSuffixToMethodName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/", ""},
		{"/config", "Config"},
		{"/post/info", "PostInfo"},
		{"post/info", "PostInfo"},
		{"post", "Post"},
	}

	fmt.Println("=== suffixToMethodName 测试 ===")
	for _, tt := range tests {
		result := suffixToMethodName(tt.input)
		status := "✓"
		if !strings.EqualFold(result, tt.expected) {
			status = "✗"
		}
		fmt.Printf("%s suffixToMethodName(%q) = %q, expected %q\n", status, tt.input, result, tt.expected)
	}
}
