package config

import "gopkg.in/yaml.v3"

type Cache struct {
	Enable bool `yaml:"enable"`
	Amount int  `yaml:"amount"`
}

func (s *Cache) UnmarshalYAML(value *yaml.Node) (err error) {

	err = value.Decode(s)

	if err == nil {
		if s.Amount == 0 {
			s.Amount = 1000
		}
	}
	return
}

// s.enDbCache = toml.Bool("cache.enable", false)
// s.cacheAmout = toml.Int("cache.amount", 1000)
