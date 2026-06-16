package ctrl

import "reflect"

type Context interface {
	RenderAs(string)
	ReqMethod() string
	SetHeader(key, value string)
	SetForceFormat(format, layout string)
	Suffix() string
	SetSuffix(string)
	Env() string
	Fill(v interface{}, fills ...bool) error
	RespondOK()
	RespondError(err error, context ...string)
	Respond(data interface{})
	NotRendered() bool
	GetPre(key string) []reflect.Value
	SetPathParams(params map[string]string)
	UnpopSuffix()
	PopSuffix() string
}

type Wrapper interface {
	UpdateRender(string, Context)
	GetRouting() []string
	MatchSuffix(string) bool
	//Get the render by user require, if required render is not allow, pass RenderNotAllowed
	GetRender() (render []string)
	//Reset the allowed renders
	SetRender([]string)
	//Call the function in object and Parse data, this function used before
	//the render prepared. So you can change function and render in here
	Parse(Context) error
	Layout() string
	TemplatePath() string
	ErrorRender() string
	AutoHidden() bool
	//Get the type of block
	String() string
}

func New(basic *Basic, block interface{}, ignoreCase bool) Wrapper {
	var ctrl Wrapper
	switch basic.typ {
	case "single":
		ctrl = &Html{basic}
	case "rest":
		ctrl = &Rest{basic}
	default:
		ctrl = &Group{basic, ignoreCase}
	}
	basic.methods.Opt = ctrl
	return ctrl
}
