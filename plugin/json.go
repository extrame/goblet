package plugin

import (
	"github.com/extrame/goblet"
)

//_JSONPlugin JSONPlugin make server default respond data as data instead of as html.
type _JSONPlugin struct {
	//some error is object and use Error() to response string, this attribute can use to control
	//whether we response Error as String by call Error() or just resonse Error Obbject
	ResponseOriginalError bool
}

type JsonErrorRender interface {
	RespondAsJson() bool
}

//JsonError mark a type as an error which should be used as Json, you can implement a type with RespondAsJson() bool function and respond true
// or just inherbit JsonError type
type JsonError struct {
}

func (j JsonError) RespondAsJson() bool {
	return true
}

//JSON standard JSON plugin
//JSON 标准JSON服务器
//Usage: goblet.Organize("server-name", plugin.JSON)
var JSON = new(_JSONPlugin)

func (p *_JSONPlugin) RespendOk(ctx *goblet.Context) {
	ctx.AddRespond("Success", true)
}

func (p *_JSONPlugin) RespondError(ctx *goblet.Context, err error, context ...string) {
	if p.ResponseOriginalError {
		ctx.AddRespond("Error", err)
	} else if je, ok := err.(JsonErrorRender); ok && je.RespondAsJson() {
		ctx.AddRespond("Error", err)
	} else {
		ctx.AddRespond("Error", err.Error())
	}
	if len(context) > 0 {
		ctx.AddRespond("Context", context)
	}
}

func (p *_JSONPlugin) DefaultRender() string {
	return "json"
}
