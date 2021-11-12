package config

type Log struct {
	File string `yaml:"file"`
}

// s.logFile = toml.String("log.file", "")
