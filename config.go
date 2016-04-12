package goblet

import (
	"fmt"
	toml "github.com/extrame/go-toml-config"
)

func (s *Server) AddConfig(name string, obj interface{}) {
	UnmarshalForm(func(tag string) []string {
		fmt.Println(name, tag)
		pText := toml.String(name+"."+tag, "")
		toml.Load()
		fmt.Println(*pText)
		if *pText != "" {
			return []string{*pText}
		} else {
			return []string{}
		}

	}, obj, true)
}
