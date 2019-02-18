package boltkv

import (
	"github.com/boltdb/bolt"
	"github.com/extrame/goblet"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

type Collection struct {
	name string
	db   *bolt.DB
}

func (c *Collection) Get(name string, pointer interface{}) (err error) {
	var bts []byte

	c.db.Update(func(tx *bolt.Tx) error {
		if b, err := tx.CreateBucketIfNotExists([]byte(c.name)); err == nil {
			bts = b.Get([]byte(name))
		}
		return nil
	})

	if len(bts) == 0 {
		return goblet.NoSuchRecord
	}
	if err = bson.Unmarshal(bts, pointer); err != nil {
		err = errors.Wrapf(err, "in bolt unmarshal(%s)", string(bts))
	}
	return err
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

func (c *Collection) Del(name string) error {

	return c.db.Update(func(tx *bolt.Tx) error {
		if b, err := tx.CreateBucketIfNotExists([]byte(c.name)); err == nil {
			return b.Delete([]byte(name))
		} else {
			return err
		}
	})
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
