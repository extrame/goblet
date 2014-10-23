package goblet

import (
	"fmt"
	_ "github.com/go-sql-driver/MySQL"
	"github.com/go-xorm/xorm"
)

var DB *xorm.Engine

func newDB(engine, user, pwd, host, name string, port int) (err error) {
	var q string
	if engine == "mysql" {
		q = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", user, pwd, host, port, name)
	}
	DB, err = xorm.NewEngine(engine, q)
	return
}
