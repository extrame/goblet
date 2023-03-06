package goblet

func TestMatcher(url string, ctrls ...interface{}) (string, string) {
	testServer := Organize("goblet-test", &StringConfiger{Content: BasicConfig})
	for _, ctrl := range ctrls {
		testServer.ControlBy(ctrl)
	}
	anchor, suffix := testServer.router.anchor.match(url, len(url))
	return anchor.opt.String(), suffix
}
