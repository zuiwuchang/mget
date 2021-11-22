package db

import (
	"errors"
	"fmt"
	"os"
	"runtime"
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

type Batch struct {
	ID  int64
	Val int64
}

func DefaultDB() *DB {
	return defaultDB
}

type DB struct {
	*bolt.DB
	Filename string
	Temp     string
	ch       chan Batch
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
	log.Trace(`open db: `, filename)
	d, e := bolt.Open(filename, 0600, &bolt.Options{Timeout: time.Second})
	if e != nil {
		e = fmt.Errorf(`open db %s-> %w`, filename, e)
		return
	}
	defaultDB = &DB{
		DB:       d,
		Filename: filename,
		Temp:     temp,
		ch:       make(chan Batch, runtime.NumCPU()*4),
	}
	go defaultDB.batch()
	result = defaultDB
	return
}
func (d *DB) Load(size, block utils.Size, modified string) error {
	log.Trace(`load db`)
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
	d.ch <- Batch{
		ID:  id,
		Val: size,
	}
	return
}
func (d *DB) batch() {
	m := make(map[int64]int64)
	for {
		d.getBatch(m)
		d.putBatch(m)
	}
}
func (d *DB) putBatch(m map[int64]int64) {
	d.Update(func(t *bolt.Tx) (e error) {
		var (
			bucket = t.Bucket(BucketTask)
		)
		for k, v := range m {
			e = bucket.Put(utils.Itob(k), utils.Itob(v))
			if e != nil {
				break
			}
		}
		return
	})
}
func (d *DB) getBatch(m map[int64]int64) {
	node := <-d.ch
	m[node.ID] = node.Val
	for {
		select {
		case node = <-d.ch:
			m[node.ID] = node.Val
		default:
			return
		}
	}
}
