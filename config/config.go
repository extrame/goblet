package config

import "gopkg.in/yaml.v3"

type Config struct {
	Basic Basic `yaml:"basic"`
	Cache Cache `yaml:"cache"`
	Log   Log   `yaml:"log"`
	Db    Db    `yaml:"db"`
}

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	type plain Config
	return value.Decode((*plain)(c))
}