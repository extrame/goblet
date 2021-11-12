package goblet

type NewPlugin interface {
	AddCfgAndInit(server *Server) error
}

//DbPwdPlugin Change the db connection password
type DbPwdPlugin interface {
	SetPwd(origin string) string
}

//DbPwdPlugin Change the db connection name
type dbUserNamePlugin interface {
	SetName(origin string) string
}

//RequestPlugin Called on the request built
type onNewRequestPlugin interface {
	OnNewRequest(*Context) error
}

type OkFuncSetter interface {
	RespendOk(*Context)
}

type ErrFuncSetter interface {
	RespondError(*Context, error, ...string)
}

type DefaultRenderSetter interface {
	DefaultRender() string
}
