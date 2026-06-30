package config

type ExampleConfig struct {
	//goblet tag指定配置项的名称，default指定默认值
	Example string `goblet:"example" default:"example"`
}
