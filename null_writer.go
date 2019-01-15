package goblet

type nullWriter struct {
}

func (n *nullWriter) Write(bts []byte) (int, error) {
	return len(bts), nil
}
