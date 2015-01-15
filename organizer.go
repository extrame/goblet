package goblet

var DefaultServer *Server

func Organize(name string) *Server {
	DefaultServer := new(Server)
	DefaultServer.Organize(name)
	return DefaultServer
}
