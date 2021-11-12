package goblet

import (
	"fmt"
	"net/url"

	"github.com/extrame/unmarshall"
	"gopkg.in/yaml.v3"
)

func fetch(node *yaml.Node) map[string]string {

	var fetched = make(map[string]string)
	for i := 0; i < len(node.Content); i++ {
		var c = node.Content[i]
		if c.Kind == yaml.ScalarNode {
			var content = node.Content[i+1]
			if content.Kind == yaml.ScalarNode {
				fetched[c.Value] = content.Value
				i = i + 1
			} else if content.Kind == yaml.SequenceNode {
				for i, sub := range content.Content {
					if sub.Kind == yaml.ScalarNode {
						fetched[fmt.Sprintf("%s[%d]", c.Value, i)] = sub.Value
					} else {
						subFetched := fetch(sub)
						for j, subFetched := range subFetched {
							fetched[fmt.Sprintf("%s[%d].%s", c.Value, i, j)] = subFetched
						}
					}
				}
				i = i + 1
			}
		}
	}

	return fetched
}

func (s *Server) AddConfig(name string, obj interface{}) error {
	var node = s.getCfg(name)
	var content = fetch(node)

	var u = unmarshall.Unmarshaller{
		ValueGetter: func(tag string) []string {
			if c, ok := content[tag]; ok {
				return []string{c}
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

func getChildNode(parent *yaml.Node, name string) *yaml.Node {
	if parent.Kind == yaml.DocumentNode && parent.Anchor == "" {
		parent = parent.Content[0]
	}
	for n, m := range parent.Content {
		if m.Kind == yaml.ScalarNode && m.Value == name && len(parent.Content) > n+1 {
			return parent.Content[n+1]
		}
	}
	return nil
}
