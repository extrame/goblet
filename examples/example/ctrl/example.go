package ctrl

import "github.com/extrame/goblet/v2"

// ExampleCtrl 控制器
// 控制器对应一个url路径，用于处理该路径下的请求
type ExampleCtrl struct {
	// 每个ctrl默认对应和其字段名相同的url路径，例如ExampleCtrl对应/example-ctrl
	// 可以通过path标签指定自定义的url路径，例如path:"/example"
	goblet.Route `path:"/example"`
}

// Get 处理GET请求
// 当用户访问/example-ctrl路径时，会调用该方法
// ctx 是上下文对象，包含了请求的所有信息，例如请求方法、请求路径、请求体等
// 可以通过ctx来获取请求参数、设置响应等
// 返回可以是两个参数，第一个参数是响应数据，第二个参数是错误信息
// 如果错误信息为nil，表示处理成功，否则表示处理失败
// 可以只返回error，表示处理失败，例如返回一个错误信息，null表示处理成功
// 这时，通过函数调用ctx.Respond(具体的响应数据)来设置具体的响应数据
func (c *ExampleCtrl) Get(ctx *goblet.Context) error {
	// 处理GET请求
	return nil
}

// 非Get/Post/PUT/DELETE等方法的请求，会按函数名调用对应的处理函数，例如调用/example/test路径时，会调用Test方法，大小写不敏感
func (c *ExampleCtrl) Test(ctx *goblet.Context) (ExampleResponse, error) {
	// 处理Test请求
	return ExampleResponse{}, nil
}

// 对/example/deep-test路径的请求，会调用DeepTest方法，大小写不敏感,但按大写对应为用-分隔的路径
func (c *ExampleCtrl) DeepTest(ctx *goblet.Context) (ExampleResponse, error) {
	// 处理Test2请求
	return ExampleResponse{}, nil
}

// 对/example/deep/test路径的请求，会调用Dee_Test方法，大小写不敏感,但按路由分段对应为用_分隔的函数名
func (c *ExampleCtrl) Deep_Test(ctx *goblet.Context) (ExampleResponse, error) {
	// 处理Test2请求
	return ExampleResponse{}, nil
}

type ExampleResponse struct {
	Count int `json:"count"`
}
