package view

import (
	"github.com/jroimartin/gocui"
)

func (v *View) Close() {
	v.g.Close()
}
func (v *View) MainLoop() error {
	return v.g.MainLoop()
}
func (v *View) Update(f func(g *gocui.Gui) error) {
	if v.g != nil {
		v.g.Update(f)
	}
}
func (v *View) SetStatus(body string) {
	if v.viewStatus != nil {
		v.viewStatus.SetBody(body)
	}
}
func (v *View) SetWorker(body string) {
	v.bodyWorker = body
	if v.viewWorker != nil {
		v.viewWorker.SetBody(body)
	}
}
