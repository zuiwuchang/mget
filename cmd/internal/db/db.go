package db

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
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
	Output   string
	ch       chan Batch
	close    chan struct{}
	wait     sync.WaitGroup
}

func OpenDB(output string) (result *DB, e error) {
	var temp, filename string
	if strings.HasSuffix(output, `.`) {
		temp = output + `tmp`
		filename = output + "db"
	} else {
		temp = output + `.tmp`
		filename = output + ".db"
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
		Output:   output,
		Temp:     temp,
		ch:       make(chan Batch, runtime.NumCPU()*4),
		close:    make(chan struct{}),
	}
	defaultDB.wait.Add(1)
	go defaultDB.batch()
	result = defaultDB
	return
}
func (d *DB) Finish() (e error) {
	e = os.Rename(d.Temp, d.Output)
	if e != nil {
		return
	}
	close(d.close)
	d.wait.Wait()
	d.Close()
	os.Remove(d.Filename)
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
	defer d.wait.Done()
	m := make(map[int64]int64)
	for {
		if d.getBatch(m) {
			break
		}
		d.putBatch(m)
		for k := range m {
			delete(m, k)
		}
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
func (d *DB) getBatch(m map[int64]int64) (exit bool) {
	var node Batch
	select {
	case node = <-d.ch:
	case <-d.close:
		exit = true
		return
	}

	m[node.ID] = node.Val
	for {
		select {
		case node = <-d.ch:
			m[node.ID] = node.Val
		case <-d.close:
			exit = true
			return
		default:
			return
		}
	}
}
