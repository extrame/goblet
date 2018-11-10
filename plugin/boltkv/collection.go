package boltkv

import (
	"github.com/boltdb/bolt"
	"gopkg.in/mgo.v2/bson"
)

type Collection struct {
	name string
	db   *bolt.DB
}

func (c *Collection) Get(name string, pointer interface{}) error {
	var bytesChan = make(chan []byte)

	go c.db.Update(func(tx *bolt.Tx) error {
		if b, err := tx.CreateBucketIfNotExists([]byte(c.name)); err == nil {
			bytesChan <- b.Get([]byte(name))
		}
		return nil
	})

	bts := <-bytesChan
	return bson.Unmarshal(bts, pointer)
}

func (c *Collection) Set(name string, pointer interface{}) error {

	if out, err := bson.Marshal(pointer); err == nil {
		return c.db.Update(func(tx *bolt.Tx) error {
			if b, err := tx.CreateBucketIfNotExists([]byte(c.name)); err == nil {
				return b.Put([]byte(name), out)
			} else {
				return err
			}
		})
	} else {
		return err
	}
}

func (c *Collection) Keys() []string {
	var bytesChan = make(chan []byte)
	var result []string

	go c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.name))
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				bytesChan <- k
				return nil
			})
		}
		bytesChan <- nil
		return nil
	})

loop:
	for {
		select {
		case bts := <-bytesChan:
			if bts == nil {
				break loop
			}
			result = append(result, string(bts))
		}
	}
	return result
}
