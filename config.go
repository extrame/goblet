package goblet

import (
	"fmt"
	"net/url"

	"github.com/extrame/unmarshall"
	"gopkg.in/yaml.v3"
)

func fetch(node *yaml.Node) map[string][]string {
	var fetched = make(map[string][]string)
	for _, c := range node.Content {
		// DocumentNode Kind = 1 << iota
		// SequenceNode
		// MappingNode
		// ScalarNode
		// AliasNode
		fetched[c.Anchor] = nil
		fmt.Println(c)
	}
	return fetched
}

func (s *Server) AddConfig(name string, obj interface{}) error {
	var node = s.cfg[name]
	var content = fetch(node)

	var u = unmarshall.Unmarshaller{
		ValueGetter: func(tag string) []string {
			if c, ok := content[tag]; ok {
				return c
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
