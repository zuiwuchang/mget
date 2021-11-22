package db

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/zuiwuchang/mget/cmd/internal/log"
	"github.com/zuiwuchang/mget/utils"
)

var defaultDB *DB

var (
	BucketMetadata = []byte(`Metadata`)
	BucketTask     = []byte(`Task`)
)

func DefaultDB() *DB {
	return defaultDB
}

type DB struct {
	*bolt.DB
	Filename string
	Temp     string
}

func OpenDB(filename string) (result *DB, e error) {
	var temp string
	if strings.HasSuffix(filename, `.`) {
		temp = filename + `tmp`
		filename += "db"
	} else {
		temp = filename + `.tmp`
		filename += ".db"
	}
	d, e := bolt.Open(filename, 0600, &bolt.Options{Timeout: time.Second})
	if e != nil {
		e = fmt.Errorf(`open db %s-> %w`, filename, e)
		return
	}
	defaultDB = &DB{
		DB:       d,
		Filename: filename,
		Temp:     temp,
	}
	result = defaultDB
	return
}
func (d *DB) Load(size, block utils.Size, modified string) error {
	return d.Update(func(t *bolt.Tx) (e error) {
		var (
			bucket = t.Bucket(BucketMetadata)
			key    = []byte(`md`)
			val    []byte
		)
		if bucket == nil {
			bucket, e = t.CreateBucket(BucketMetadata)
			if e != nil {
				return
			}
			md := &Metadata{Size: size,
				Block:    block,
				Modified: modified,
			}
			val, e = md.Marshal()
			if e != nil {
				return
			}
			e = bucket.Put(key, val)
			if e != nil {
				return
			}
			_, e = t.CreateBucket(BucketTask)
			if e != nil {
				return
			}
			e = d.truncate(d.Temp, int64(size))
			if e != nil {
				return
			}
			log.Info(`new metadata`)
		} else {
			val = bucket.Get(key)
			md := &Metadata{}
			e = md.Unmarshal(val)
			if e != nil {
				return
			}
			if md.Size == size && md.Block == block && md.Modified == modified {
				log.Info(`metadata matched`)
				return
			}
			e = errors.New(`metadata not matched`)
			return
		}
		return
	})
}
func (d *DB) truncate(filename string, size int64) (e error) {
	f, e := os.Create(filename)
	if e != nil {
		return
	}
	e = f.Truncate(size)
	f.Close()
	if e != nil {
		os.Remove(filename)
	}
	return
}

func (d *DB) GetSize(id int64) (size utils.Size, e error) {
	e = d.View(func(t *bolt.Tx) (e error) {
		var (
			bucket = t.Bucket(BucketTask)
			key    = utils.Itob(id)
		)
		size = utils.Size(utils.Btoi(bucket.Get(key)))
		return
	})
	return
}
func (d *DB) SetSize(id, size int64) (e error) {
	e = d.Update(func(t *bolt.Tx) (e error) {
		var (
			bucket = t.Bucket(BucketTask)
			key    = utils.Itob(id)
			val    = utils.Itob(size)
		)
		e = bucket.Put(key, val)
		return
	})
	return
}
