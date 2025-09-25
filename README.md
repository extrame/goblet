Goblet Golang Http 框架 / Goblet Golang HTTP Framework

# 概述 / Overview
Goblet是一款基于Go语言开发的HTTP框架，他的第一版已经有11年的历史了，整体设计上参考了Ruby on Rails的设计理念，目标不是提供轻量化框架，而是提供一个功能丰富、易于上手的解决方案。

Goblet is a HTTP framework developed in Go language. Its first version has a history of 11 years. The overall design is inspired by Ruby on Rails, aiming not to provide a lightweight framework, but a feature-rich and easy-to-use solution.

# v2版本设计目标 / v2 Design Goals

在11年后，Goblet的v2版本将重新设计，目标是提供一个更现代、更灵活和易于使用的框架。以下是v2版本的几个主要设计目标：

After 11 years, Goblet v2 will be redesigned with the goal of providing a more modern, flexible and easy-to-use framework. Here are the main design goals for v2:

* 切换至Gorm，原来设计基于xorm，v2将切换到Gorm，这将带来更好的性能和更丰富的功能。
  * Switch to Gorm, the original design was based on xorm, v2 will switch to Gorm, which will bring better performance and richer features.
* 引入更灵活的路由系统，支持参数化路由和中间件。
  * Introduce a more flexible routing system that supports parameterized routes and middleware.
* 默认以JSON作为数据交换格式，并提供灵活的配置选项。原来优先Html模板，v2将改为优先JSON。
  * Use JSON as the default data exchange format and provide flexible configuration options. Originally prioritized HTML templates, v2 will prioritize JSON.
* 从logrus切换到slog
  * Switch from logrus to slog
* 提供完善的文档和示例，帮助开发者快速上手，也帮助AI更好地理解框架的使用。
  * Provide complete documentation and examples to help developers get started quickly and also help AI better understand the framework usage.
* 提供更清晰的命名规范和代码结构，提高可读性和维护性。
  * Provide clearer naming conventions and code structure to improve readability and maintainability.
* 如果可能的话，提升性能。
  * Improve performance if possible.
* 使用标准化的form，json解析库和tag标注模式
  * Use standardized form, json parsing libraries and tag annotation patterns.

  # 创建一个新的项目 / Create a new project

  ```golang
  import "github.com/extrame/goblet"

  func main() {
    var server = goblet.Organize("project_name")
    server.ControlBy(&MyController{})
    server.Run()
  }
  ```
