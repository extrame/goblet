package goblet

var defaultServer *Server

//Organize 生成一个goblet服务器
func Organize(name string, plugins ...Plugin) *Server {
	defaultServer := new(Server)
	defaultServer.Organize(name, plugins)
	return defaultServer
}
