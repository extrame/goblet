package json

import "github.com/extrame/goblet"

type _Plugin struct {
}

var Plugin = new(_Plugin)

func (p *_Plugin) RespendOk(ctx *goblet.Context) {
	ctx.AddRespond("Success", true)
}

func (p *_Plugin) RespondError(ctx *goblet.Context, err error, context ...string) {
	ctx.AddRespond("Error", err.Error())
	if len(context) > 0 {
		ctx.AddRespond("Context", context)
	}
}

func (p *_Plugin) DefaultRender() string {
	return "json"
}
