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
	switch basic.typ {
	case "single":
		return &Html{basic}
	case "rest":
		return &Rest{basic}
	case "group":
		return &Group{basic, ignoreCase}
	}

	for i := 0; i < basic.block.Type().NumMethod(); i++ {
		mtd := basic.block.Type().Method(i)
		switch mtd.Name {
		case "Get", "Post":
			return &Html{basic}
		case "Read", "ReadMany", "Delete", "DeleteMany", "Update", "UpdateMany", "New", "Create", "Edit":
			return &Rest{basic}
		case "Init":
			continue
		default:
			return &Group{basic, ignoreCase}
		}
	}
	return &Group{basic, ignoreCase}
}
