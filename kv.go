package goblet

import (
	"errors"

	"github.com/sirupsen/logrus"
)

var NoSuchRecord = errors.New("no such record")

//用于获取数值的接口
type Kv interface {
	Get(name string, pointer interface{}) error
	Set(name string, pointer interface{}) error
	Del(name string) error
	Keys() []string
}

//用户指定KV驱动的接口
type KvDriver interface {
	Collection(string) Kv //specified the table name and return the collection
}

func (s *Server) KV(name string) Kv {
	if s.kv != nil {
		return s.kv.Collection(name)
	} else {
		logrus.Errorln("not specified kv driver, please specified in server.Organize func")
		return nil
	}
}
