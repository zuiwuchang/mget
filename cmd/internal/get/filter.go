package get

import (
	"time"

	"github.com/jroimartin/gocui"
)

type Filter struct {
	g     *gocui.Gui
	ch    chan func(*gocui.Gui) error
	close chan struct{}
}

func NewFilter(g *gocui.Gui) *Filter {
	f := &Filter{
		g:     g,
		ch:    make(chan func(*gocui.Gui) error, 1),
		close: make(chan struct{}),
	}
	return f
}
func (filter *Filter) Close() {
	close(filter.close)
}
func (filter *Filter) Serve() {
	for {
		select {
		case f := <-filter.ch:
			if filter.send(f) {
				return
			}
		case <-filter.close:
			return
		}
	}
}
func (filter *Filter) send(f func(*gocui.Gui) error) (closed bool) {
	var (
		t        *time.Timer
		duration = time.Millisecond * 100
	)

	for {
		if t == nil {
			t = time.NewTimer(duration)
		} else {
			t.Reset(duration)
		}

		select {
		case <-t.C:
			filter.g.Update(f)
			return
		case <-filter.close:
			closed = true
			return
		case f = <-filter.ch:
			if !t.Stop() {
				<-t.C
			}
		}
	}
}
func (filter *Filter) Update(f func(*gocui.Gui) error) {
	for {
		select {
		case <-filter.close:
			return
		case filter.ch <- f:
			return
		default:
		}
		select {
		case <-filter.close:
			return
		case <-filter.ch:
		default:
		}
	}
}
