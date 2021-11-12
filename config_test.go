package goblet

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/unmarshall"
	"gopkg.in/yaml.v3"
)

func TestConfigSubStruct(t *testing.T) {
	var obj struct {
		Type   string `goblet:"type,t1"`
		Detail []struct {
			Name string `goblet:"name,here"`
		} `goblet:"detail"`
	}

	var node = make(map[string]*yaml.Node)

	err := yaml.NewDecoder(strings.NewReader(`
basic:
  www_root: ./www
type:
  test: t2
  array:
    - name: 1
      sex: 2
    - name: 2
      sex: 3
  `)).Decode(&node)

	fmt.Println(node, err)

	var m = fetch(node["type"])

	fmt.Println(m)

	var u = unmarshall.Unmarshaller{
		ValueGetter: func(tag string) []string {

			return []string{}

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

func TestConfigSubArray(t *testing.T) {
	var obj struct {
		Type   string `goblet:"type,t1"`
		Detail []struct {
			Name string `goblet:"name,here"`
		} `goblet:"detail"`
	}

	err := toml.Parse("./test/test_array.config")

	if err != nil {
		t.Fatal(err)
	}

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
