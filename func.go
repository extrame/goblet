package goblet

func (s *Server) AddFunc(name string, fn interface{}) {
	s.funcs = append(s.funcs, Fn{name, fn})
}
