package goblet

func (s *Server) AddFunc(name string, obj interface{}) {
	s.funcs[name] = obj
}
