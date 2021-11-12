package config

import "gopkg.in/yaml.v3"

type Cache struct {
	Enable bool `yaml:"enable"`
	Amount int  `yaml:"amount"`
}

func (s *Cache) UnmarshalYAML(value *yaml.Node) (err error) {
	type plain Cache
	err = value.Decode((*plain)(s))

	if err == nil {
		if s.Amount == 0 {
			s.Amount = 1000
		}
	}
	return
}
