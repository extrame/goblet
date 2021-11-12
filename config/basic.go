package config

import "gopkg.in/yaml.v3"

type Basic struct {
	WwwRoot         string `yaml:"www_root"`
	Port            int    `yaml:"port"`
	ReadT0          int    `yaml:"read_timeout"`
	WriteT0         int    `yaml:"write_timeout"`
	PublicDir       string `yaml:"public_dir"`
	UploadsDir      string `yaml:"uploads_dir"`
	IgnoreUrlCase   bool   `yaml:"ignore_url_case"`
	HashSecret      string `yaml:"secret"`
	Env             string `yaml:"env"`
	DbEngine        string `yaml:"db_engine"`
	Version         string `yaml:"version"`
	HttpsEnable     bool   `yaml:"https"`
	HttpsCertFile   string `yaml:"https_certfile"`
	HttpsKey        string `yaml:"https_key"`
	DefaultType     string `yaml:"default_type"`
	EnableKeepAlive bool   `yaml:"keep_alive"`
}

func (s *Basic) UnmarshalYAML(value *yaml.Node) (err error) {

	s.WwwRoot = "./www"
	s.Port = 8080
	s.ReadT0 = 30
	s.WriteT0 = 30
	s.PublicDir = "public"
	s.UploadsDir = "./uploads"
	s.IgnoreUrlCase = true
	s.HashSecret = "a238974b2378c39021d23g43"
	s.Env = ProductEnv
	s.DbEngine = "mysql"
	s.EnableKeepAlive = true
	type plain Basic
	return value.Decode((*plain)(s))
}
