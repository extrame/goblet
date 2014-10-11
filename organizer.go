package goblet

import ()

func Organize(name string) *Server {
	s := new(Server)
	s.Organize(name)
	return s
}
