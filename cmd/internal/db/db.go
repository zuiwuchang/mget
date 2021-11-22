package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

var defaultDB *DB

func DefaultDB() *DB {
	return defaultDB
}

type DB struct {
	*bolt.DB
	Filename string
}

func OpenDB(filename string) (result *DB, e error) {
	if strings.HasSuffix(filename, `.`) {
		filename += "db"
	} else {
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
	}
	result = defaultDB
	return
}
