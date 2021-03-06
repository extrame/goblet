package goblet

import (
	"net/url"

	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/unmarshall"
)

func (s *Server) AddConfig(name string, obj interface{}) error {
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
		ValuesGetter: func(prefix string) url.Values {
			return make(url.Values)
		},
		TagConcatter: func(prefix string, tag string) string {
			return prefix + "." + tag
		},
		AutoFill: true,
	}
	return u.Unmarshall(obj)
}
