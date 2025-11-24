package goblet

func TestMatcher(url string, ctrls ...interface{}) (string, string, map[string]string) {
	testServer := Organize("goblet-test", &StringConfiger{Content: BasicConfig})
	for _, ctrl := range ctrls {
		testServer.ControlBy(ctrl)
	}
	anchor, suffix, params := testServer.router.anchor.Match(url, len(url))
	return anchor.Opt.String(), suffix, params
}
