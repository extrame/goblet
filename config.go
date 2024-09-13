package goblet

import (
	"fmt"

	"github.com/creasty/defaults"
	myyaml "github.com/extrame/unmarshall/yaml"
	"github.com/sirupsen/logrus"
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

func (s *Server) AddConfig(name string, obj interface{}, tagName ...string) error {
	err := defaults.Set(obj)
	if err != nil {
		logrus.Debug(err)
	}
	node, _ := myyaml.GetChildNode(s.cfg, name)
	var _tagName string
	if len(tagName) > 0 {
		_tagName = tagName[0]
	} else {
		_tagName = "goblet"
	}
	return myyaml.UnmarshalNode(node, obj, _tagName)
}

func (s *Server) getCfg(name string) *yaml.Node {
	node, err := myyaml.GetChildNode(s.cfg, name)
	if err == nil {
		return node
	} else {
		return new(yaml.Node)
	}
}
