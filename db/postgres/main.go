package postgres

import (
	"github.com/extrame/goblet"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	goblet.RegisterDB("mysql", func(s string) gorm.Dialector {
		return postgres.Open(s)
	})
}
