package goblet

import "gorm.io/gorm"

var dialectorCreators = make(map[string]func(string) gorm.Dialector)

func RegisterDB(name string, fn func(string) gorm.Dialector) {
	dialectorCreators[name] = fn
}
