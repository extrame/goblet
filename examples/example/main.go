package main

import (
	"github.com/extrame/goblet/v2"
	"github.com/extrame/goblet/v2/examples/example/ctrl"
	"github.com/extrame/goblet/v2/examples/example/model"
)

// main 函数启动Goblet应用
func main() {
	// 使用Organize函数组织应用，指定应用名称为"example"，名称对应会要求加载运行目录的example.conf文件
	// 作为应用的配置文件
	app := goblet.Organize("example")

	// 注册ExampleCtrl控制器, 控制器对应一个url路径
	app.ControlBy(&ctrl.ExampleCtrl{})

	// 注册自动同步的模型，会使用gorm自动同步模型
	app.AddModel(&model.ExampleModel{})

	// 启动应用，等待应用终止
	err := app.Run()
	// err包含应用启动时的错误信息，以及终止应用的错误信息
	if err != nil {
		panic(err)
	}
}
