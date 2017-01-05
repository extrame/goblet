package goblet

type Plugin interface {
	ParseConfig(prefix string) error
	Init() error
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
