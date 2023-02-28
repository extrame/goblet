package goblet

func TestMatcher(ctrl interface{}, url string) (string, bool) {
	testServer := Organize("goblet-test", &StringConfiger{Content: BasicConfig})
	testServer.ControlBy(ctrl)
	anchor, suffix := testServer.router.anchor.match(url, len(url))
	return suffix, anchor != nil
}
