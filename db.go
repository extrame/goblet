package goblet

import (
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
	// _ "github.com/mattn/go-oci8"
)

var DB *xorm.Engine

func ResetDB() error {
	DB.Close()
	return DefaultServer.connectDB()
}
