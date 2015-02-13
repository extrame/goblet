package goblet

type Plugin interface {
	ParseConfig() error
	Init() error
}
