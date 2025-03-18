package sqlserver

import (
	"github.com/extrame/goblet"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func init() {
	goblet.RegisterDB("mysql", func(s string) gorm.Dialector {
		return sqlserver.Open(s)
	})
}
