package plugin

import "github.com/extrame/goblet"

type _JSONPlugin struct {
}

var JSON = new(_JSONPlugin)

func (p *_JSONPlugin) RespendOk(ctx *goblet.Context) {
	ctx.AddRespond("Success", true)
}

func (p *_JSONPlugin) RespondError(ctx *goblet.Context, err error, context ...string) {
	ctx.AddRespond("Error", err.Error())
	if len(context) > 0 {
		ctx.AddRespond("Context", context)
	}
}

func (p *_JSONPlugin) DefaultRender() string {
	return "json"
}
