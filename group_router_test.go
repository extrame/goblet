package goblet

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/extrame/goblet/v2/internal/ctrl"
	"github.com/extrame/goblet/v2/render"
)

// GroupMatchController 采用在 GroupController 字段 tag 上直接声明 path 与 render 的写法。
//
// GroupController 会对控制器路由之后的后缀做方法匹配：
//   - /api/group/index      -> Index()
//   - /api/group/show       -> Show()
//   - /api/group/get-detail -> GetDetail()
//   - 未匹配到任何后缀时     -> 回退到 HTTP 方法（GET -> Get()）
type GroupMatchController struct {
	GroupController `path:"/api/group" render:"json"`
}

func (c *GroupMatchController) Index(ctx *Context) string {
	return "index"
}

func (c *GroupMatchController) Show(ctx *Context) string {
	return "show"
}

func (c *GroupMatchController) GetDetail(ctx *Context) string {
	return "get-detail"
}

func (c *GroupMatchController) Get(ctx *Context) string {
	return "get"
}

func (c *GroupMatchController) Create(ctx *Context) string {
	return "create"
}

// GroupMatchLegacyController 采用原有写法（Route / Render 独立字段），用于验证兼容性。
type GroupMatchLegacyController struct {
	Route          `path:"/api/legacy"`
	Render         `json`
	GroupController
}

func (c *GroupMatchLegacyController) Index(ctx *Context) string {
	return "legacy-index"
}

func (c *GroupMatchLegacyController) Get(ctx *Context) string {
	return "legacy-get"
}

type groupMatchCase struct {
	name     string
	url      string
	expected string
}

func runGroupCases(t *testing.T, s *Server, cases []groupMatchCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("url %s: expected status 200, got %d (body: %s)", tt.url, rec.Code, rec.Body.String())
			}

			var resp StandardErrorOrData
			if err := json.Unmarshal([]byte(strings.TrimSpace(rec.Body.String())), &resp); err != nil {
				t.Fatalf("url %s: response is not JSON: %s (err: %v)", tt.url, rec.Body.String(), err)
			}
			if resp.Data != tt.expected {
				t.Errorf("url %s: expected data %q, got %q", tt.url, tt.expected, resp.Data)
			}
		})
	}
}

func TestGroupControllerURLMatch(t *testing.T) {
	server := Organize("group-test", &StringConfiger{Content: BasicConfig}, new(render.JsonRender))
	server.ControlBy(&GroupMatchController{})

	runGroupCases(t, server, []groupMatchCase{
		{name: "root-get", url: "/api/group", expected: "get"},
		{name: "index-suffix", url: "/api/group/index", expected: "index"},
		{name: "show-suffix", url: "/api/group/show", expected: "show"},
		{name: "kebab-case-suffix", url: "/api/group/get-detail", expected: "get-detail"},
		// 后缀未匹配到方法时，回退到 HTTP 方法 Get()
		{name: "fallback-to-http-method", url: "/api/group/unknown", expected: "get"},
		// 疑似 Create 方法无法匹配的复现用例
		{name: "create-suffix", url: "/api/group/create", expected: "create"},
	})
}

// TestGroupControllerURLMatchLegacy 验证旧的 Route / Render 字段写法仍然可用。
func TestGroupControllerURLMatchLegacy(t *testing.T) {
	server := Organize("group-legacy-test", &StringConfiger{Content: BasicConfig}, new(render.JsonRender))
	server.ControlBy(&GroupMatchLegacyController{})

	runGroupCases(t, server, []groupMatchCase{
		{name: "root-get", url: "/api/legacy", expected: "legacy-get"},
		{name: "index-suffix", url: "/api/legacy/index", expected: "legacy-index"},
	})
}

// RestMatchController 采用在 RestController 字段 tag 上直接声明 path 与 render 的写法。
type RestMatchController struct {
	RestController `path:"/api/rest" render:"json"`
}

func (c *RestMatchController) Index(ctx *Context) string {
	return "rest-index"
}

// HttpMatchController 采用在 HttpMethodController 字段 tag 上直接声明 path 与 render 的写法。
type HttpMatchController struct {
	HttpMethodController `path:"/api/http" render:"json"`
}

// TestControllerTagsParsedByDetectOption 验证三种控制器类型都能从字段 tag 解析出 path 与 render。
func TestControllerTagsParsedByDetectOption(t *testing.T) {
	cases := []struct {
		name       string
		block      interface{}
		wantRoute  string
		wantRender string
	}{
		{name: "group", block: &GroupMatchController{}, wantRoute: "/api/group", wantRender: "json"},
		{name: "rest", block: &RestMatchController{}, wantRoute: "/api/rest", wantRender: "json"},
		{name: "http", block: &HttpMatchController{}, wantRoute: "/api/http", wantRender: "json"},
	}
	for _, tc := range cases {
		basic, _ := ctrl.DetectOption(tc.block, &Server{})
		routing := basic.GetRouting()
		if len(routing) != 1 || routing[0] != tc.wantRoute {
			t.Errorf("%s: routing = %v, want %q", tc.name, routing, tc.wantRoute)
		}
		renders := basic.GetRender()
		if len(renders) != 1 || renders[0] != tc.wantRender {
			t.Errorf("%s: render = %v, want %q", tc.name, renders, tc.wantRender)
		}
	}
}

// TestRestControllerURLMatch 端到端验证 RestController 的新 tag 写法可用。
func TestRestControllerURLMatch(t *testing.T) {
	server := Organize("rest-tag-test", &StringConfiger{Content: BasicConfig}, new(render.JsonRender))
	server.ControlBy(&RestMatchController{})

	runGroupCases(t, server, []groupMatchCase{
		{name: "rest-index", url: "/api/rest", expected: "rest-index"},
	})
}

// TestGroupCreateMatchByPost 验证 POST 请求下 Create 方法同样能被匹配（更贴近 CRUD 写操作）。
func TestGroupCreateMatchByPost(t *testing.T) {
	server := Organize("group-create-post", &StringConfiger{Content: BasicConfig}, new(render.JsonRender))
	server.ControlBy(&GroupMatchController{})

	req := httptest.NewRequest(http.MethodPost, "/api/group/create", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/group/create: expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp StandardErrorOrData
	if err := json.Unmarshal([]byte(strings.TrimSpace(rec.Body.String())), &resp); err != nil {
		t.Fatalf("response is not JSON: %s (err: %v)", rec.Body.String(), err)
	}
	if resp.Data != "create" {
		t.Errorf("expected data %q, got %q", "create", resp.Data)
	}
}

// PrefixAppController 与 PrefixAppConfigController 用于验证前缀冲突场景。
// 期望 /api/app/config 锚定 /api/app，而不是被 /api/app-config 抢占。
type PrefixAppController struct {
	GroupController `path:"/api/app" render:"json"`
}

func (c *PrefixAppController) Config(ctx *Context) string {
	return "app-config-method"
}

type PrefixAppConfigController struct {
	GroupController `path:"/api/app-config" render:"json"`
}

func (c *PrefixAppConfigController) Config(ctx *Context) string {
	return "appconfig-config-method"
}

func TestGroupPrefixConflictRoute(t *testing.T) {
	server := Organize("prefix-conflict-test", &StringConfiger{Content: BasicConfig}, new(render.JsonRender))
	// 挂载顺序：先 /api/app，后 /api/app-config（与用户描述一致）
	server.ControlBy(&PrefixAppController{})
	server.ControlBy(&PrefixAppConfigController{})

	runGroupCases(t, server, []groupMatchCase{
		{name: "app-should-win", url: "/api/app/config", expected: "app-config-method"},
		{name: "appconfig-config", url: "/api/app-config/config", expected: "appconfig-config-method"},
	})
}