package goblet

func (s *Server) AddFunc(name string, obj func(*Context) interface{}) {
	s.funcs[name] = obj
}
