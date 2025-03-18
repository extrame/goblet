package mysql

import (
	"github.com/extrame/goblet"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	goblet.RegisterDB("mysql", func(s string) gorm.Dialector {
		return mysql.Open(s)
	})
}
