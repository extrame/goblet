package goblet

var DefaultServer *Server

func Organize(name string, plugins ...Plugin) *Server {
	DefaultServer := new(Server)
	DefaultServer.Organize(name, plugins)
	return DefaultServer
}
