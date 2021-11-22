package log

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/zuiwuchang/mget/widget"
)

var defaultLog = NewLog(100)
var logWriter net.Conn

func logStr(str string) {
	if logWriter == nil {
		var e error
		logWriter, e = net.Dial(`tcp`, `127.0.0.1:7000`)
		if e != nil {
			return
		}
	}
	_, e := logWriter.Write([]byte(str + "\n"))
	if e != nil {
		logWriter.Close()
		logWriter = nil
	}
}

type Log struct {
	widget       *widget.Widget
	strs         []string
	offset, size int
	body         string
	m            sync.Mutex
}

func NewLog(size int) *Log {
	if size < 1 {
		size = 10
	}

	return &Log{
		strs: make([]string, size),
	}
}
func (l *Log) Display() bool {
	l.m.Lock()
	ok := l.widget != nil
	l.m.Unlock()
	return ok
}
func (l *Log) Layout() (e error) {
	l.m.Lock()
	defer l.m.Unlock()
	if l.widget != nil {
		e = l.widget.Layout()
	}
	return
}

func (l *Log) Toggle(f func(string) *widget.Widget) {
	l.m.Lock()
	defer l.m.Unlock()
	if l.widget == nil {
		l.widget = f(l.body)
	} else {
		l.widget.DeleteView()
		l.widget = nil
	}
}
func (l *Log) Tag(tag string, a ...interface{}) {
	l.m.Lock()
	defer l.m.Unlock()
	var str string
	if len(a) != 0 {
		str = fmt.Sprint(a...)
	}
	str = time.Now().Format(`2006/01/02 15:04:05`) + ` [` + tag + `] ` + str
	logStr(str)
	l.push(str)
	if l.widget != nil {
		l.widget.SetBodyAndScroll(l.body, true)
	}
}
func (l *Log) Tagf(tag, format string, a ...interface{}) {
	l.m.Lock()
	defer l.m.Unlock()
	if len(a) != 0 {
		format = fmt.Sprintf(format, a...)
	}
	str := time.Now().Format(`2006/01/02 15:04:05`) + ` [` + tag + `] ` + format
	logStr(str)
	l.push(str)
	if l.widget != nil {
		l.widget.SetBodyAndScroll(l.body, true)
	}
}
func (l *Log) Errorf(format string, a ...interface{}) {
	l.Tagf(`error`, format, a...)
}
func (l *Log) Infof(format string, a ...interface{}) {
	l.Tagf(`info`, format, a...)
}
func (l *Log) Tracef(format string, a ...interface{}) {
	l.Tagf(`trace`, format, a...)
}
func (l *Log) Error(a ...interface{}) {
	l.Tag(`error`, a...)
}
func (l *Log) Info(a ...interface{}) {
	l.Tag(`info`, a...)
}
func (l *Log) Trace(a ...interface{}) {
	l.Tag(`trace`, a...)
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
