package plugin

type DelimSetter struct {
	prefix string
	suffix string
}

func (d *DelimSetter) SetDelim() [2]string {
	return [2]string{d.prefix, d.suffix}
}

func NewDelimSetter(strs ...string) *DelimSetter {
	prefix, suffix := "{{", "}}"
	if len(strs) > 0 {
		prefix = strs[0]
	}
	if len(strs) > 1 {
		suffix = strs[1]
	}
	return &DelimSetter{
		prefix: prefix,
		suffix: suffix,
	}
}
