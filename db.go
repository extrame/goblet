package goblet

import (
	"xorm.io/xorm"
)

var DB *xorm.Engine

func ResetDB() error {
	DB.Close()
	return DefaultServer.connectDB()
}
