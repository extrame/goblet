package plugin

func SetConfigSuffix(suffix string) *_SetConfigSuffixPlugin {
	return &_SetConfigSuffixPlugin{
		suffix: suffix,
	}
}

type _SetConfigSuffixPlugin struct {
	suffix string
}

func (s *_SetConfigSuffixPlugin) GetConfigSuffix() string {
	return s.suffix
}
