package goblet

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/extrame/goblet/internal/ctrl"
	"github.com/extrame/goblet/internal/matcher"
)

// 基准测试：参数式URL vs 普通URL匹配效率对比
func BenchmarkParamVsNormalURL(b *testing.B) {
	// 设置测试路由
	anchor := matcher.Root(ctrl.NewStatic())

	// 添加普通路由
	anchor.Add("/api/users/list", &ctrl.Html{&ctrl.Basic{Name: "UserList"}})
	anchor.Add("/api/users/profile", &ctrl.Html{&ctrl.Basic{Name: "UserProfile"}})
	anchor.Add("/api/posts/recent", &ctrl.Html{&ctrl.Basic{Name: "RecentPosts"}})
	anchor.Add("/api/posts/popular", &ctrl.Html{&ctrl.Basic{Name: "PopularPosts"}})
	anchor.Add("/api/products/search", &ctrl.Html{&ctrl.Basic{Name: "ProductSearch"}})

	// 添加参数路由
	anchor.Add("/api/users/:id", &ctrl.Html{&ctrl.Basic{Name: "UserDetail"}})
	anchor.Add("/api/posts/:postId", &ctrl.Html{&ctrl.Basic{Name: "PostDetail"}})
	anchor.Add("/api/products/:category/:productId", &ctrl.Html{&ctrl.Basic{Name: "ProductDetail"}})

	b.Run("NormalURL_Match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			anchor.Match("/api/users/list", len("/api/users/list"))
		}
	})

	b.Run("NormalURL_Match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			anchor.Match("/api/products/search", len("/api/products/search"))
		}
	})

	b.Run("ParamURL_Match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			anchor.Match("/api/users/123", len("/api/users/123"))
		}
	})

	b.Run("ComplexParamURL_Match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			anchor.Match("/api/products/electronics/456", len("/api/products/electronics/456"))
		}
	})
}

// 基准测试：混合场景下的参数路由 vs 普通路由
func BenchmarkMixedScenario(b *testing.B) {
	anchor := matcher.Root(ctrl.NewStatic())

	// 混合添加普通路由和参数路由，模拟真实场景
	routes := []struct {
		path    string
		isParam bool
	}{
		{"/", false},
		{"/home", false},
		{"/about", false},
		{"/contact", false},
		{"/api/users", false},
		{"/api/users/:id", true},
		{"/api/users/:id/profile", true},
		{"/api/posts", false},
		{"/api/posts/:postId", true},
		{"/api/posts/:postId/comments", true},
		{"/api/products", false},
		{"/api/products/:category", true},
		{"/api/products/:category/:productId", true},
		{"/api/search", false},
		{"/api/search/:query", true},
	}

	for _, route := range routes {
		anchor.Add(route.path, &ctrl.Html{&ctrl.Basic{Name: "Controller"}})
	}

	// 测试路径集合
	testPaths := []string{
		"/home",
		"/api/users",
		"/api/users/123",
		"/api/posts",
		"/api/posts/456",
		"/api/products/electronics/789",
	}

	b.Run("Mixed_NormalPriority", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := testPaths[i%len(testPaths)]
			anchor.Match(path, len(path))
		}
	})

	b.Run("Mixed_ParamPriority", func(b *testing.B) {
		// 更多参数路由测试
		paramPaths := []string{
			"/api/users/111",
			"/api/posts/222/comments",
			"/api/products/books/333",
			"/api/search/golang",
		}
		for i := 0; i < b.N; i++ {
			path := paramPaths[i%len(paramPaths)]
			anchor.Match(path, len(path))
		}
	})
}

// 基准测试：参数提取开销分析
func BenchmarkParamExtractionOverhead(b *testing.B) {
	anchor := matcher.Root(ctrl.NewStatic())

	// 添加单参数路由
	anchor.Add("/user/:id", &ctrl.Html{&ctrl.Basic{Name: "UserDetail"}})

	// 添加多参数路由
	anchor.Add("/shop/:category/:product/:variant", &ctrl.Html{&ctrl.Basic{Name: "ProductDetail"}})

	b.Run("SingleParam_Extraction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, params := anchor.Match("/user/12345", len("/user/12345"))
			if params != nil && params["id"] != "12345" {
				b.Fatal("参数提取失败")
			}
		}
	})

	b.Run("MultiParam_Extraction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, params := anchor.Match("/shop/electronics/laptop/16gb", len("/shop/electronics/laptop/16gb"))
			if params != nil {
				if params["category"] != "electronics" ||
					params["product"] != "laptop" ||
					params["variant"] != "16gb" {
					b.Fatal("多参数提取失败")
				}
			}
		}
	})
}

// 基准测试：不同参数格式的匹配效率
func BenchmarkParamFormatComparison(b *testing.B) {
	anchor := matcher.Root(ctrl.NewStatic())

	// 不同参数格式的路由
	anchor.Add("/user/:id", &ctrl.Html{&ctrl.Basic{Name: "Format1"}})      // :param
	anchor.Add("/post/{id}", &ctrl.Html{&ctrl.Basic{Name: "Format2"}})     // {param}
	anchor.Add("/product/<id>", &ctrl.Html{&ctrl.Basic{Name: "Format3"}})  // <param>
	anchor.Add("/category/[id]", &ctrl.Html{&ctrl.Basic{Name: "Format4"}}) // [param]

	b.Run("ColonFormat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			anchor.Match("/user/123", len("/user/123"))
		}
	})

	b.Run("BraceFormat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			anchor.Match("/post/456", len("/post/456"))
		}
	})

	b.Run("AngleFormat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			anchor.Match("/product/789", len("/product/789"))
		}
	})

	b.Run("BracketFormat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			anchor.Match("/category/abc", len("/category/abc"))
		}
	})
}

// 基准测试：参数路由 vs 正则匹配 - 参数路由性能对比
func BenchmarkParamRouteVsRegex(b *testing.B) {
	// 设置路由匹配器
	anchor := matcher.Root(ctrl.NewStatic())
	anchor.Add("/api/:id/test", &ctrl.Html{&ctrl.Basic{Name: "BenchController1"}})
	anchor.Add("/api/:id/profile", &ctrl.Html{&ctrl.Basic{Name: "BenchController2"}})
	anchor.Add("/api/:id/settings", &ctrl.Html{&ctrl.Basic{Name: "BenchController3"}})
	anchor.Add("/api/:id/friends", &ctrl.Html{&ctrl.Basic{Name: "BenchController4"}})
	anchor.Add("/api/:id/messages", &ctrl.Html{&ctrl.Basic{Name: "BenchController5"}})

	// 设置正则表达式匹配器
	patterns := []string{
		`^/api/([^/]+)/test$`,
		`^/api/([^/]+)/profile$`,
		`^/api/([^/]+)/settings$`,
		`^/api/([^/]+)/friends$`,
		`^/api/([^/]+)/messages$`,
	}
	regexps := make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		regexps[i] = regexp.MustCompile(pattern)
	}

	b.Run("ParamRoute_Goblet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = anchor.Match("/api/123/test", len("/api/123/test"))
		}
	})

	b.Run("ParamRoute_Regex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := "/api/123/test"
			for _, re := range regexps {
				if re.MatchString(path) {
					break
				}
			}
		}
	})
}

// 基准测试：大规模参数路由匹配性能
func BenchmarkLargeScaleParamRoutes(b *testing.B) {
	anchor := matcher.Root(ctrl.NewStatic())

	// 添加大量参数路由
	for i := 0; i < 100; i++ {
		path := fmt.Sprintf("/api/resource%d/:id", i)
		anchor.Add(path, &ctrl.Html{&ctrl.Basic{Name: fmt.Sprintf("Resource%d", i)}})
	}

	// 添加普通路由作为对比
	for i := 0; i < 100; i++ {
		path := fmt.Sprintf("/api/static%d/fixed", i)
		anchor.Add(path, &ctrl.Html{&ctrl.Basic{Name: fmt.Sprintf("Static%d", i)}})
	}

	b.Run("LargeScale_ParamRoutes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resourceId := i % 100
			path := fmt.Sprintf("/api/resource%d/%d", resourceId, i)
			anchor.Match(path, len(path))
		}
	})

	b.Run("LargeScale_StaticRoutes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			resourceId := i % 100
			path := fmt.Sprintf("/api/static%d/fixed", resourceId)
			anchor.Match(path, len(path))
		}
	})
}
