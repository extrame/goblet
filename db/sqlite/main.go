package sqlite

import (
	"github.com/extrame/goblet"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	goblet.RegisterDB("mysql", func(s string) gorm.Dialector {
		return sqlite.Open(s)
	})
}
