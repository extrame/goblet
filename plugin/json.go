package plugin

import "github.com/extrame/goblet"

type _JSONPlugin struct {
	//some error is object and use Error() to response string, this attribute can use to control
	//whether we response Error as String by call Error() or just resonse Error Obbject
	ResponseOriginalError bool
}

var JSON = new(_JSONPlugin)

func (p *_JSONPlugin) RespendOk(ctx *goblet.Context) {
	ctx.AddRespond("Success", true)
}

func (p *_JSONPlugin) RespondError(ctx *goblet.Context, err error, context ...string) {
	if p.ResponseOriginalError {
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
