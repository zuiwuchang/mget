package db

import (
	"bytes"
	"encoding/gob"

	"github.com/zuiwuchang/mget/utils"
)

type Metadata struct {
	Size     utils.Size
	Block    utils.Size
	Modified string
}

func (m *Metadata) Marshal() ([]byte, error) {
	var w bytes.Buffer
	e := gob.NewEncoder(&w).Encode(m)
	if e != nil {
		return nil, e
	}
	return w.Bytes(), nil
}
func (m *Metadata) Unmarshal(b []byte) error {
	r := bytes.NewReader(b)
	return gob.NewDecoder(r).Decode(m)
}
