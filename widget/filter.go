package widget

import (
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
	ch := make(chan int)
	for {
		select {
		case f := <-filter.ch:
			if filter.send(ch, f) {
				return
			}
		case <-filter.close:
			return
		}
	}
}
func (filter *Filter) send(ch chan int, f func(*gocui.Gui) error) (closed bool) {
	filter.g.Update(func(g *gocui.Gui) error {
		select {
		case ch <- 1:
		case <-filter.close:
		}
		return f(g)
	})
	select {
	case <-ch:
	case <-filter.close:
		closed = true
	}
	return
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
