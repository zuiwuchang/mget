package db

import (
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

var db *DB

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
		return
	}
	db = &DB{
		DB:       d,
		Filename: filename,
	}
	result = db
	return
}
