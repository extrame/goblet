package goblet

import (
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/MySQL"
	"github.com/go-xorm/xorm"
	// _ "github.com/mattn/go-oci8"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *xorm.Engine

func newDB(engine, user, pwd, host, name string, port int, con_to int, ka_intv int) (err error) {
	var q string
	if engine == "mysql" {
		q = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", user, pwd, host, port, name)
	} else if engine == "oci8" {
		q = fmt.Sprintf("%s/%s@%s:%d/%s", user, pwd, host, port, name)
	} else if engine == "mssql" {
		q = fmt.Sprintf("Server=%s;Database=%s;User ID=%s;Password=%s;connection timeout=%d;keepAlive=%d", host, name, user, pwd, con_to, ka_intv)
	} else if engine == "sqlite3" {
		if info, err := os.Stat(host + ".db"); err == nil {
			if info.IsDir() {
				return fmt.Errorf("If you want to use sqlite3, please set db.host as rw file")
			}
		}
		q = host + ".db"
	} else if engine == "none" {
		return
	}
	DB, err = xorm.NewEngine(engine, q)
	return
}

func ResetDB() error {
	DB.Close()
	return DefaultServer.connectDB()
}
