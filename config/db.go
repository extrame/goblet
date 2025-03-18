package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"

	// "gorm.io/driver/mysql"
	// "gorm.io/driver/postgres"
	// "gorm.io/driver/sqlite"
	// "gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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

func (d *Db) New(engine string, dialectorCreator func(string) gorm.Dialector) (db *gorm.DB, err error) {
	var dialector gorm.Dialector
	var dsn string

	switch engine {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			d.User, d.Pwd, d.Host, d.Port, d.Name)

	case "postgres":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			d.Host, d.Port, d.User, d.Pwd, d.Name)

	case "mssql":
		dsn = fmt.Sprintf("Server=%s;Database=%s;User Id=%s;Password=%s;connection timeout=%d",
			d.Host, d.Name, d.User, d.Pwd, d.TO)

	case "sqlite3", "sqlite":
		dsn = d.Host + ".db"
		if info, err := os.Stat(dsn); err == nil && info.IsDir() {
			return nil, fmt.Errorf("if you want to use sqlite3, please set db.host as rw file")
		}

	case "none":
		return nil, NoDbDriver

	default:
		return nil, fmt.Errorf("unsupported db type:%s,supported:[mysql,postgres,mssql,sqlite3,sqlite,none]", engine)
	}

	slog.Info("connecting to DB", "db type", engine)
	dialector = dialectorCreator(dsn)

	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   d.Prefix,
			SingularTable: true,
		},
	}

	return gorm.Open(dialector, config)
}

func (s *Db) UnmarshalYAML(value *yaml.Node) (err error) {
	s.Port = 3306
	s.TO = 30
	type plain Db
	return value.Decode((*plain)(s))
}
