package goblet

import (
	"fmt"
	"net/url"
	"testing"

	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/unmarshall"
)

func TestConfigSubStruct(t *testing.T) {
	var obj struct {
		Type   string `goblet:"type,t1"`
		Detail struct {
			Name string `goblet:"name,here"`
		} `goblet:"detail"`
	}

	toml.Parse("./test/test.config")

	var u = unmarshall.Unmarshaller{
		ValueGetter: func(tag string) []string {
			pText := toml.String("test"+"."+tag, "")
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
	u.Unmarshall(&obj)
	fmt.Println(obj)
}
