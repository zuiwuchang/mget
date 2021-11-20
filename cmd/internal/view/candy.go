package view

import "github.com/jroimartin/gocui"

func (v *View) Close() {
	v.filterStatus.Close()
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
	v.filterStatus.Update(func(g *gocui.Gui) error {
		v.viewStatus.SetBody(body)
		return nil
	})
}
