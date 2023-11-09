package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

var NoDbDriver = errors.New("no db driver for this server")

type Db struct {
	Host       string `yaml:"host"`
	User       string `yaml:"user"`
	Pwd        string `yaml:"password"`
	Name       string `yaml:"name"`
	Port       int    `yaml:"port"`
	TO         int    `yaml:"connect_timeout"`
	KaInterval int    `yaml:"ka_interval"`
	Prefix     string `yaml:"prefix"`
}

func (d *Db) New(engine string) (db *xorm.Engine, err error) {

	var q string
	if engine == "mysql" {
		q = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", d.User, d.Pwd, d.Host, d.Port, d.Name)
	} else if engine == "postgres" {
		q = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", d.Host, d.Port, d.User, d.Pwd, d.Name)
	} else if engine == "oci8" {
		q = fmt.Sprintf("%s/%s@%s:%d/%s", d.User, d.Pwd, d.Host, d.Port, d.Name)
	} else if engine == "mssql" {
		q = fmt.Sprintf("Server=%s;Database=%s;User ID=%s;Password=%s;connection timeout=%d;keepAlive=%d", d.Host, d.Name, d.User, d.Pwd, d.TO, d.KaInterval)
	} else if engine == "sqlite3" || engine == "sqlite" {
		if info, err := os.Stat(d.Host + ".db"); err == nil {
			if info.IsDir() {
				return nil, fmt.Errorf("If you want to use sqlite3, please set db.host as rw file")
			}
		}
		q = d.Host + ".db"
	} else if engine == "none" {
		return nil, NoDbDriver
	} else {
		return nil, fmt.Errorf("unsupported db type:%s,supported:[mysql,oci8,mssql,sqlite3,sqlite,none]", engine)
	}
	logrus.WithField("db type", engine).WithField("url", q).Infoln("connect to DB")
	db, err = xorm.NewEngine(engine, q)
	if err == nil {
		if d.Prefix != "" {
			db.SetTableMapper(names.NewPrefixMapper(names.SnakeMapper{}, d.Prefix))
		}
	}
	return db, err
}

func (s *Db) UnmarshalYAML(value *yaml.Node) (err error) {
	s.Port = 3306
	s.TO = 30
	type plain Db
	return value.Decode((*plain)(s))
}
