package plugin

type _SilenceSetter struct {
	m map[string]bool
}

func (s *_SilenceSetter) SetSilenceUrls() map[string]bool {
	return s.m
}

func Silent(url ...string) *_SilenceSetter {
	var m = make(map[string]bool)
	for _, v := range url {
		m[v] = true
	}
	return &_SilenceSetter{
		m: m,
	}
}
