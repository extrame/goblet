package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
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

func (d *Db) New(engine string) (db *gorm.DB, err error) {
	var dialector gorm.Dialector

	switch engine {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			d.User, d.Pwd, d.Host, d.Port, d.Name)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			d.Host, d.Port, d.User, d.Pwd, d.Name)
		dialector = postgres.Open(dsn)
	case "sqlite3", "sqlite":
		if info, err := os.Stat(d.Host + ".db"); err == nil && info.IsDir() {
			return nil, fmt.Errorf("if you want to use sqlite3, please set db.host as rw file")
		}
		dialector = sqlite.Open(d.Host + ".db")
	case "mssql":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&connection+timeout=%d",
			d.User, d.Pwd, d.Host, d.Port, d.Name, d.TO)
		dialector = sqlserver.Open(dsn)
	case "none":
		return nil, NoDbDriver
	default:
		return nil, fmt.Errorf("unsupported db type:%s,supported:[mysql,postgres,mssql,sqlite3,sqlite,none]", engine)
	}

	logrus.WithField("db type", engine).Infoln("connecting to DB")
	db, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *Db) UnmarshalYAML(value *yaml.Node) (err error) {
	s.Port = 3306
	s.TO = 30
	type plain Db
	return value.Decode((*plain)(s))
}
