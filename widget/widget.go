package widget

import (
	"errors"
	"fmt"

	"github.com/jroimartin/gocui"
)

type Widget struct {
	gui    *gocui.Gui
	name   string
	body   string
	layout Layout
	view   *gocui.View

	scroll bool
}

func NewWidget(gui *gocui.Gui, name, body string, layout Layout) (widget *Widget, e error) {
	if name == `` {
		e = errors.New(`invalid name`)
		return
	}
	_, e = gui.View(name)
	if e == nil {
		e = fmt.Errorf(`view already exists: %s`, name)
		return
	} else if e != gocui.ErrUnknownView {
		return
	} else {
		e = nil
	}
	x, y, with, height := layout.Rect()
	view, e := gui.SetView(name, x, y, x+with, y+height)
	if e != nil {
		if e != gocui.ErrUnknownView {
			return
		}
		e = nil
	}

	widget = &Widget{
		gui:    gui,
		name:   name,
		body:   body,
		layout: layout,
		view:   view,
	}
	return
}
func (w *Widget) View() *gocui.View {
	return w.view
}
func (w *Widget) GUI() *gocui.Gui {
	return w.gui
}
func (w *Widget) Layout() (e error) {
	x, y, with, height := w.layout.Rect()
	v, e := w.gui.SetView(w.name, x, y, x+with, y+height)
	if e != nil {
		if e != gocui.ErrUnknownView {
			return
		}
		e = nil
	}
	v.Clear()
	fmt.Fprintf(v, w.body)
	return
}

func (w *Widget) SetBody(text string) {
	w.gui.Update(func(g *gocui.Gui) error {
		w.body = text
		return nil
	})
}
func (w *Widget) UnsafeSetBody(text string) {
	w.body = text
}
func (w *Widget) SetBodyAndScroll(text string, bottom bool) {
	w.gui.Update(func(g *gocui.Gui) error {
		w.body = text
		if bottom {
			w.ScrollBottom()
		} else {
			w.ScrollTop()
		}
		return nil
	})
}
func (w *Widget) Body() string {
	return w.body
}

func (w *Widget) Update(f func(*gocui.Gui) error) {
	w.gui.Update(f)
}
func (w *Widget) DeleteView() error {
	w.DeleteKeybindings()
	return w.gui.DeleteView(w.name)
}
func (w *Widget) DeleteKeybinding(key interface{}, mod gocui.Modifier) error {
	return w.gui.DeleteKeybinding(w.name, key, mod)
}
func (w *Widget) DeleteKeybindings() {
	w.gui.DeleteKeybindings(w.name)
}
func (w *Widget) SetKeybinding(key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	return w.gui.SetKeybinding(w.name, key, mod, handler)
}
func (w *Widget) Scroll() bool {
	return w.scroll
}
func (w *Widget) EnableScroll(scroll bool) {
	if scroll {
		if w.scroll {
			return
		}
		w.scroll = true
		w.SetKeybinding(gocui.MouseWheelUp, gocui.ModNone, w.scrollUp)
		w.SetKeybinding(gocui.MouseWheelDown, gocui.ModNone, w.scrollDown)
		w.SetKeybinding(gocui.KeyArrowUp, gocui.ModNone, w.scrollUp)
		w.SetKeybinding(gocui.KeyArrowDown, gocui.ModNone, w.scrollDown)
		w.SetKeybinding(gocui.KeyPgup, gocui.ModNone, w.scrollPageUp)
		w.SetKeybinding(gocui.KeyPgdn, gocui.ModNone, w.scrollPageDown)
	} else {
		if !w.scroll {
			return
		}
		w.scroll = false
		w.DeleteKeybinding(gocui.MouseWheelUp, gocui.ModNone)
		w.DeleteKeybinding(gocui.MouseWheelDown, gocui.ModNone)
		w.DeleteKeybinding(gocui.KeyArrowUp, gocui.ModNone)
		w.DeleteKeybinding(gocui.KeyArrowDown, gocui.ModNone)
		w.DeleteKeybinding(gocui.KeyPgup, gocui.ModNone)
		w.DeleteKeybinding(gocui.KeyPgdn, gocui.ModNone)
	}
}
func (w *Widget) scrollUp(g *gocui.Gui, v *gocui.View) (e error) {
	x, y := v.Origin()
	if y > 0 {
		e = v.SetOrigin(x, y-1)
	}
	return
}
func (w *Widget) scrollPageUp(g *gocui.Gui, v *gocui.View) (e error) {
	x, y := v.Origin()
	if y > 0 {
		_, sy := v.Size()
		y -= sy
		if y < 0 {
			y = 0
		}
		e = v.SetOrigin(x, y)
	}
	return
}
func (w *Widget) scrollDown(g *gocui.Gui, v *gocui.View) (e error) {
	count := len(v.ViewBufferLines())
	_, sy := v.Size()
	max := count - sy
	x, y := v.Origin()
	if y >= max {
		return
	}
	e = v.SetOrigin(x, y+1)
	if e != nil {
		return
	}
	return
}
func (w *Widget) scrollPageDown(g *gocui.Gui, v *gocui.View) (e error) {
	count := len(v.ViewBufferLines())
	_, sy := v.Size()
	max := count - sy
	x, y := v.Origin()
	if y >= max {
		return
	}
	y += sy
	if y > max {
		y = max
	}
	e = v.SetOrigin(x, y)
	if e != nil {
		return
	}
	return
}
func (w *Widget) ScrollTop() {
	w.gui.Update(func(g *gocui.Gui) error {
		return w.scrollTop()
	})
}
func (w *Widget) scrollTop() (e error) {
	v := w.view
	x, y := v.Origin()
	if y > 0 {
		e = v.SetOrigin(x, 0)
	}
	return
}
func (w *Widget) ScrollBottom() {
	w.gui.Update(func(g *gocui.Gui) error {
		return w.scrollBottom()
	})
}
func (w *Widget) scrollBottom() (e error) {
	v := w.view
	count := len(v.ViewBufferLines())
	_, sy := v.Size()
	max := count - sy
	x, y := v.Origin()
	if y >= max {
		return
	}
	e = v.SetOrigin(x, max)
	if e != nil {
		return
	}
	return
}
