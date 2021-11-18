package get

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type Widget struct {
	name string
	x, y int
	w, h int
	body string

	disabled, hide bool
}

func (w *Widget) Layout(g *gocui.Gui) (e error) {
	v, e := g.SetView(w.name, w.x, w.y, w.x+w.w, w.y+w.h)
	if e != nil {
		if e != gocui.ErrUnknownView {
			return
		}
		fmt.Fprint(v, w.body)
	}
	return
}
