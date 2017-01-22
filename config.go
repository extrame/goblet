package goblet

import (
	toml "github.com/extrame/go-toml-config"
)

func (s *Server) AddConfig(name string, obj interface{}) {
	UnmarshalForm(func(tag string) []string {
		pText := toml.String(name+"."+tag, "")
		toml.Load()
		if *pText != "" {
			return []string{*pText}
		} else {
			return []string{}
		}

	}, nil, obj, true)
}
