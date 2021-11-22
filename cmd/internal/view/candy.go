package view

import (
	"github.com/jroimartin/gocui"
	"github.com/zuiwuchang/mget/cmd/internal/log"
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
	log.Info(`SetStatus `)
	defer log.Info(`SetStatus exit`)

	if v.viewStatus != nil {
		v.viewStatus.SetBody(body)
	}
}
func (v *View) SetWorker(body string) {
	log.Info(`SetWorker `)
	defer log.Info(`SetWorker exit`)

	v.bodyWorker = body
	if v.viewWorker != nil {
		v.viewWorker.SetBody(body)
	}
}
