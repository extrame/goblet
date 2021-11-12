package goblet

type Option struct {
	DB string
}

func (o Option) overlay(opts []Option) {
	for _, v := range opts {
		if v.DB != "" {
			o.DB = v.DB
		}
	}
}
