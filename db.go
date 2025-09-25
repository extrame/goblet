package goblet

import (
	"gorm.io/gorm"
)

var DB *gorm.DB

func ResetDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.Close()
	return DefaultServer.connectDB()
}
