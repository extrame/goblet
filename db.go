package goblet

import (
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/MySQL"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

var DB *xorm.Engine

func newDB(engine, user, pwd, host, name string, port int) (err error) {
	var q string
	if engine == "mysql" {
		q = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", user, pwd, host, port, name)
	} else if engine == "mssql" {
		q = fmt.Sprintf("Server=%s;Database=%s;User ID=%s;Password=%s;", host, name, user, pwd)
	} else if engine == "sqlite3" {
		if info, err := os.Stat(host); err == nil {
			if info.IsDir() {
				return fmt.Errorf("If you want to use sqlite3, please set db.host as rw file")
			}
		}
		q = host
	}
	DB, err = xorm.NewEngine(engine, q)
	return
}
