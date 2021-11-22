package db

import (
	"github.com/zuiwuchang/mget/utils"
)

type Task struct {
	ID     int64
	Offset utils.Size
	Num    utils.Size
}
