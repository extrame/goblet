package boltkv

import (
	"log"

	"github.com/boltdb/bolt"
	"github.com/extrame/goblet"
)

type BackEnd struct {
	db *bolt.DB
}

type BackEndConfig struct {
	File string `goblet:file`
}

var defaultConfig BackEndConfig

func (b *BackEnd) init(path string, option *bolt.Options) error {
	db, err := bolt.Open(path, 0600, option)
	if err != nil {
		log.Panicf("backend: cannot open database at %s (%v)", path, err)
	} else {
		b.db = db
	}
	return err
}

func (b *BackEnd) AddCfgAndInit(server *goblet.Server) error {

	server.AddConfig("bolt", &defaultConfig)

	if defaultConfig.File == "" {
		defaultConfig.File = server.Name + ".db"
	}

	return b.init(defaultConfig.File, nil)
}

func (b *BackEnd) Collection(name string) goblet.Kv {
	return &Collection{
		name: name,
		db:   b.db,
	}
}
