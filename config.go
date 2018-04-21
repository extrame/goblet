package goblet

import (
	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/unmarshall"
)

func (s *Server) AddConfig(name string, obj interface{}) {
	var u = unmarshall.Unmarshaller{
		ValueGetter: func(tag string) []string {
			pText := toml.String(name+"."+tag, "")
			toml.Load()
			if *pText != "" {
				return []string{*pText}
			} else {
				return []string{}
			}

		},
		TagConcatter: func(prefix string, tag string) string {
			return prefix + "." + tag
		},
		AutoFill: true,
	}
	u.Unmarshall(obj)
}
