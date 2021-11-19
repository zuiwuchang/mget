package get

import (
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/zuiwuchang/mget/widget"
)

const (
	ViewClose  = `close`
	ViewConf   = `conf`
	ViewStatus = `status`
	ViewWorker = `worker`
	ViewLog    = `log`
	ViewMenu   = `menu`
)
const (
	MenuWidth    = 25
	ConfHeight   = 4
	StatusHeight = 2
	LogHeight    = 4
	LogCount     = 100
)

type GUI struct {
	g          *gocui.Gui
	m          *Manager
	viewConf   *widget.Widget
	viewStatus *widget.Widget
	viewWorker *widget.Widget
	viewMenu   *widget.Widget
}

func NewGUI(conf *Configure, m *Manager) *GUI {
	return &GUI{
		m: m,
	}
}
func (gui *GUI) Init() (e error) {
	g, e := gocui.NewGui(gocui.OutputNormal)
	if e != nil {
		return
	}
	gui.g = g
	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen
	g.Mouse = true

	g.SetManagerFunc(gui.layout)
	gui.toggleLog(g, nil)
	gui.toggleMenu(g, nil)
	if e = g.SetKeybinding(``, gocui.KeyCtrlC, gocui.ModNone, gui.quit); e != nil {
		return
	} else if e = g.SetKeybinding(``, 'm', gocui.ModNone, gui.toggleMenu); e != nil {
		return
	} else if e = g.SetKeybinding(``, 'l', gocui.ModNone, gui.toggleLog); e != nil {
		return
	} else if e = g.SetKeybinding(``, 'w', gocui.ModNone, gui.increaseWorker); e != nil {
		return
	} else if e = g.SetKeybinding(``, 's', gocui.ModNone, gui.reduceWorker); e != nil {
		return
	} else if e = g.SetKeybinding(``, gocui.MouseLeft, gocui.ModNone, gui.current); e != nil {
		return
	}

	gui.g = g
	return
}

func (gui *GUI) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
func (gui *GUI) toggleMenu(g *gocui.Gui, _ *gocui.View) error {
	g.Update(func(g *gocui.Gui) error {
		if gui.viewMenu == nil {
			var e error
			gui.viewMenu, e = widget.NewWidget(g, ViewMenu,
				`w: increase worker
s: reduce worker
m: toggle menu display
l: toggle log display`,
				widget.NewLayout(func() (x int, y int, w int, h int) {
					_, h = g.Size()
					h--
					w = MenuWidth
					return
				}),
			)
			if e != nil {
				defaultLog.Push(`error`, `display menu: `+e.Error())
				return nil
			}
			view := gui.viewMenu.View()
			view.Title = ViewMenu
			g.SetKeybinding(ViewMenu, gocui.MouseLeft, gocui.ModNone, gui.clickMenu)
		} else {
			gui.viewMenu.DeleteView()
			gui.viewMenu = nil
		}
		return nil
	})
	return nil
}
func (gui *GUI) toggleLog(g *gocui.Gui, _ *gocui.View) error {
	g.Update(func(g *gocui.Gui) error {
		if defaultLog.Widget == nil {
			var e error
			defaultLog.Widget, e = widget.NewWidget(g, ViewLog,
				defaultLog.body,
				widget.NewLayout(func() (x int, y int, w int, h int) {
					h = LogHeight
					w, y = g.Size()
					y -= h + 1
					w--
					if gui.viewMenu != nil {
						x += MenuWidth + 1
						w -= MenuWidth + 1
					}
					return
				}),
			)
			if e != nil {
				defaultLog.Push(`error`, `display log: `+e.Error())
				return nil
			}
			view := defaultLog.Widget.View()
			view.Wrap = true
			view.Title = ViewLog
			defaultLog.Widget.EnableScroll(true)

		} else {
			defaultLog.Widget.DeleteView()
			defaultLog.Widget = nil
		}
		return nil
	})
	return nil
}

func (gui *GUI) current(g *gocui.Gui, v *gocui.View) error {
	name := v.Name()
	if _, e := g.SetCurrentView(name); e != nil {
		defaultLog.Push(`error`, `SetCurrentView %s: %v`, name, e)
		return nil
	}
	return nil
}

func (gui *GUI) layout(g *gocui.Gui) (e error) {
	if gui.viewConf == nil {
		gui.viewConf, e = widget.NewWidget(g, ViewConf,
			strings.TrimRight(gui.m.conf.String(), "\n"),
			widget.NewLayout(func() (x int, y int, w int, h int) {
				h = ConfHeight
				w, _ = g.Size()
				w--
				if gui.viewMenu != nil {
					x += MenuWidth + 1
					w -= MenuWidth + 1
				}
				return
			}),
		)
		if e == nil {
			view := gui.viewConf.View()
			view.Wrap = true
			view.Title = ViewConf
			gui.viewConf.EnableScroll(true)
		} else {
			defaultLog.Push(`error`, `new widget `+ViewConf+`: `+e.Error())
		}
	}
	if gui.viewConf != nil {
		gui.viewConf.Layout()
	}
	if gui.viewStatus == nil {
		gui.viewStatus, e = widget.NewWidget(g, ViewStatus,
			``,
			widget.NewLayout(func() (x int, y int, w int, h int) {
				h = StatusHeight
				y = ConfHeight + 1
				w, _ = g.Size()
				w--
				if gui.viewMenu != nil {
					x += MenuWidth + 1
					w -= MenuWidth + 1
				}
				return
			}),
		)
		if e == nil {
			view := gui.viewStatus.View()
			view.Wrap = true
			view.Title = ViewStatus
			gui.viewStatus.EnableScroll(true)
		} else {
			defaultLog.Push(`error`, `new widget `+ViewStatus+`: `+e.Error())
		}
	}
	if gui.viewStatus != nil {
		gui.viewStatus.Layout()
	}
	if gui.viewWorker == nil {
		gui.viewWorker, e = widget.NewWidget(g, ViewWorker,
			``,
			widget.NewLayout(func() (x int, y int, w int, h int) {
				y = ConfHeight + 1 + StatusHeight + 1
				w, h = g.Size()
				w--
				if gui.viewMenu != nil {
					x += MenuWidth + 1
					w -= MenuWidth + 1
				}
				h -= y + 1
				if defaultLog.Widget != nil {
					h -= LogHeight + 1
				}
				return
			}),
		)
		if e == nil {
			view := gui.viewWorker.View()
			view.Wrap = true
			view.Title = ViewWorker
			gui.viewWorker.EnableScroll(true)
		} else {
			defaultLog.Push(`error`, `new widget `+ViewWorker+`: `+e.Error())
		}
	}
	if gui.viewWorker != nil {
		gui.viewWorker.Layout()
	}

	if gui.viewMenu != nil {
		gui.viewMenu.Layout()
	}
	if defaultLog.Widget != nil {
		defaultLog.Widget.Layout()
	}
	return
}
func (gui *GUI) increaseWorker(g *gocui.Gui, _ *gocui.View) error {
	gui.m.Increase()
	return nil
}
func (gui *GUI) reduceWorker(g *gocui.Gui, _ *gocui.View) error {
	gui.m.Reduce()
	return nil
}

func (gui *GUI) clickMenu(g *gocui.Gui, v *gocui.View) error {
	_, y := v.Cursor()
	str, e := v.Line(y)
	if e != nil || len(str) < 1 {
		return nil
	}
	switch str[:1] {
	case "m":
		gui.toggleMenu(g, nil)
	case "l":
		gui.toggleLog(g, nil)
	case "w":
		gui.m.Increase()
	case "s":
		gui.m.Reduce()
	}
	return nil
}
