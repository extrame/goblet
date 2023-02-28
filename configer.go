package goblet

import (
	"bytes"
	"flag"
	"io"
	"os"
	"path/filepath"
)

type Configer interface {
	GetConfigSource(s *Server) (io.Reader, error)
}

type YamlConfiger struct {
}

func (c *YamlConfiger) GetConfigSource(s *Server) (io.Reader, error) {
	if s.cfgFileSuffix == "" {
		s.cfgFileSuffix = "conf"
	}
	flag.StringVar(&s.ConfigFile, "config", "./"+s.Name+"."+s.cfgFileSuffix, "设置配置文件的路径")
	flag.Parse()

	s.ConfigFile = filepath.FromSlash(s.ConfigFile)
	return os.Open(s.ConfigFile)
}

//StringConfiger is a configer that read config from a string
type StringConfiger struct {
	Content string
}

const BasicConfig = `
basic:
  env: development
  db_engine: none
`

func (c *StringConfiger) GetConfigSource(s *Server) (io.Reader, error) {
	return bytes.NewReader([]byte(c.Content)), nil
}
