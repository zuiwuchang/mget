package get

import (
	"fmt"
	"time"

	"github.com/zuiwuchang/mget/widget"
)

var defaultLog = NewLog(LogCount)

type Log struct {
	Widget       *widget.Widget
	strs         []string
	offset, size int
	body         string
}

func NewLog(size int) *Log {
	if size < 1 {
		size = 10
	}
	return &Log{
		strs: make([]string, size),
	}
}
func (l *Log) Push(tag, format string, a ...interface{}) {
	if len(a) != 0 {
		format = fmt.Sprintf(format, a...)
	}
	str := time.Now().Format(`2006/01/02 15:04:05`) + ` [` + tag + `] ` + format
	l.push(str)
	if l.Widget != nil {
		l.Widget.SetBodyAndScroll(l.body, true)
	}
}
func (l *Log) push(str string) {
	strs := l.strs
	if l.size < len(l.strs) {
		if l.size == 0 {
			l.body = str
		} else {
			l.body += "\n" + str
		}
		strs[l.size] = str
		l.size++
	} else {
		top := strs[l.offset]
		l.body = l.body[len(top)+len("\n"):] + "\n" + str
		l.offset++
		if l.offset == l.size {
			l.offset = 0
		}
		index := (l.offset + l.size - 1) % l.size
		strs[index] = str
	}
}
