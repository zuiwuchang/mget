package widget

import "github.com/jroimartin/gocui"

type Log struct {
	*Widget
	strs         []string
	offset, size int
}

func NewLog(widget *Widget, size int) *Log {
	if size < 1 {
		size = 10
	}
	return &Log{
		Widget: widget,
		strs:   make([]string, size),
	}
}
func (l *Log) Push(tag, text string) {
	str := `[` + tag + `] ` + text
	l.Update(func(g *gocui.Gui) error {
		l.push(str)
		return nil
	})
}
func (l *Log) push(str string) {
	strs := l.strs
	if l.size < len(l.strs) {
		l.body += "\n" + str
		strs[l.size] = str
		l.size++
	} else {
		top := strs[l.offset]
		l.body = l.body[len(top)+len("\n"):] + "\n" + str
		l.offset++
		index := (l.offset + l.size) % l.size
		strs[index] = str
	}
}
