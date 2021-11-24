package view

import (
	"runtime"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/zuiwuchang/mget/cmd/internal/log"
	"github.com/zuiwuchang/mget/widget"
)

type Rely interface {
	Increase()
	Reduce()
	ConfigureView() string
	ASCII() bool
}

const (
	ViewConf   = `conf`
	ViewStatus = `status`
	ViewWorker = `worker`
	ViewLog    = `log`
	ViewMenu   = `menu`
)

var (
	Views = []string{ViewConf,
		ViewStatus,
		ViewWorker,
		ViewLog,
		ViewMenu,
	}
)

const (
	MenuWidth    = 25
	ConfHeight   = 5
	StatusHeight = 2
	LogHeight    = 4
	LogCount     = 100
)

type View struct {
	g            *gocui.Gui
	rely         Rely
	viewConf     *widget.Widget
	viewStatus   *widget.Widget
	viewWorker   *widget.Widget
	viewMenu     *widget.Widget
	xMenu, yMenu int
	bodyWorker   string
}

func New(rely Rely) *View {
	return &View{
		rely: rely,
	}
}

func (v *View) Init() (e error) {
	g, e := gocui.NewGui(gocui.OutputNormal)
	if e != nil {
		return
	}
	defer func() {
		if e != nil {
			g.Close()
		}
	}()
	v.g = g
	if runtime.GOOS == `windows` {
		g.ASCII = true
	}
	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen
	g.Mouse = true

	g.SetManagerFunc(v.layout)
	v.toggleLog(g, nil)
	v.toggleMenu(g, nil)
	if e = g.SetKeybinding(``, gocui.KeyCtrlC, gocui.ModNone, v.quit); e != nil {
		return
	} else if e = g.SetKeybinding(``, 'm', gocui.ModNone, v.toggleMenu); e != nil {
		return
	} else if e = g.SetKeybinding(``, 'l', gocui.ModNone, v.toggleLog); e != nil {
		return
	} else if e = g.SetKeybinding(``, 'w', gocui.ModNone, v.increaseWorker); e != nil {
		return
	} else if e = g.SetKeybinding(``, 's', gocui.ModNone, v.reduceWorker); e != nil {
		return
	} else if e = g.SetKeybinding(``, gocui.KeyTab, gocui.ModNone, v.tab); e != nil {
		return
	} else if e = g.SetKeybinding(``, gocui.MouseLeft, gocui.ModNone, v.current); e != nil {
		return
	}
	return
}

func (v *View) quit(g *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}
func (v *View) toggleMenu(g *gocui.Gui, _ *gocui.View) error {
	g.Update(func(g *gocui.Gui) error {
		if v.viewMenu == nil {
			var e error
			v.viewMenu, e = widget.NewWidget(g, ViewMenu,
				`w: increase worker
s: reduce worker
m: toggle menu display
l: toggle log display`,
				widget.NewLayout(func() (x int, y int, w int, h int) {
					x, _ = g.Size()
					x -= MenuWidth + 1
					h = ConfHeight
					w = MenuWidth
					return
				}),
			)
			if e != nil {
				log.Error(`display menu: `, e)
				return nil
			}
			view := v.viewMenu.View()
			view.Highlight = true
			view.Title = ViewMenu
			view.SetCursor(v.xMenu, v.yMenu)
			g.SetKeybinding(ViewMenu, gocui.MouseLeft, gocui.ModNone, v.clickMenu)
			g.SetKeybinding(ViewMenu, gocui.KeyArrowUp, gocui.ModNone, v.upMenu)
			g.SetKeybinding(ViewMenu, gocui.KeyArrowDown, gocui.ModNone, v.downMenu)
			g.SetKeybinding(ViewMenu, gocui.KeyEnter, gocui.ModNone, v.enterMenu)
		} else {
			v.xMenu, v.yMenu = v.viewMenu.View().Cursor()
			v.viewMenu.DeleteView()
			v.viewMenu = nil
		}
		return nil
	})
	return nil
}
func (v *View) toggleLog(g *gocui.Gui, _ *gocui.View) error {
	g.Update(func(g *gocui.Gui) error {
		log.Toggle(func(body string) (w *widget.Widget) {
			w, e := widget.NewWidget(g, ViewLog,
				body,
				widget.NewLayout(func() (x int, y int, w int, h int) {
					h = LogHeight
					w, y = g.Size()
					y -= h + 1
					w--
					return
				}),
			)
			if e != nil {
				log.Error(`display log: `, e)
				return
			}
			view := w.View()
			view.Wrap = true
			view.Title = ViewLog
			w.EnableScroll(true)
			return
		})
		return nil
	})
	return nil
}

func (v *View) current(g *gocui.Gui, view *gocui.View) error {
	name := view.Name()
	if _, e := g.SetCurrentView(name); e != nil {
		log.Errorf(`SetCurrentView %s: %v`, name, e)
		return nil
	}
	return nil
}
func (v *View) getName(i int) (name string) {
	for {
		i = i % len(Views)
		name = Views[i]
		if name == ViewLog && !log.Display() {
			i++
			continue
		} else if name == ViewMenu && v.viewMenu == nil {
			i++
			continue
		}
		break
	}
	return
}
func (v *View) tab(g *gocui.Gui, _ *gocui.View) error {
	view := g.CurrentView()
	var current string
	if view == nil {
		current = Views[0]
	} else {
		current = view.Name()
		for i, name := range Views {
			if current == name {
				i++
				current = v.getName(i)
				break
			}
		}
	}

	if _, e := g.SetCurrentView(current); e != nil {
		log.Errorf(`SetCurrentView %s: %v`, Views[0], e)
	}
	return nil
}
func (v *View) layout(g *gocui.Gui) (e error) {
	if v.viewConf == nil {
		v.viewConf, e = widget.NewWidget(g, ViewConf,
			strings.TrimRight(v.rely.ConfigureView(), "\n"),
			widget.NewLayout(func() (x int, y int, w int, h int) {
				h = ConfHeight
				w, _ = g.Size()
				w--
				if v.viewMenu != nil {
					w -= MenuWidth + 1
				}
				return
			}),
		)
		if e == nil {
			view := v.viewConf.View()
			view.Wrap = true
			view.Title = ViewConf
			v.viewConf.EnableScroll(true)
		} else {
			log.Errorf(`new widget %s: %v`, ViewConf, e)
		}
	}
	if v.viewConf != nil {
		v.viewConf.Layout()
	}
	if v.viewStatus == nil {
		v.viewStatus, e = widget.NewWidget(g, ViewStatus,
			``,
			widget.NewLayout(func() (x int, y int, w int, h int) {
				h = StatusHeight
				y = ConfHeight + 1
				w, _ = g.Size()
				w--
				return
			}),
		)
		if e == nil {
			view := v.viewStatus.View()
			view.Wrap = true
			view.Title = ViewStatus
			v.viewStatus.EnableScroll(true)
		} else {
			log.Errorf(`new widget %s: %v`, ViewStatus, e)
		}
	}
	if v.viewStatus != nil {
		v.viewStatus.Layout()
	}
	if v.viewWorker == nil {
		v.viewWorker, e = widget.NewWidget(g, ViewWorker,
			v.bodyWorker,
			widget.NewLayout(func() (x int, y int, w int, h int) {
				y = ConfHeight + 1 + StatusHeight + 1
				w, h = g.Size()
				w--
				h -= y + 1
				if log.Display() {
					h -= LogHeight + 1
				}
				return
			}),
		)
		if e == nil {
			view := v.viewWorker.View()
			view.Wrap = true
			view.Title = ViewWorker
			v.viewWorker.EnableScroll(true)
		} else {
			log.Errorf(`new widget %s: %v`, ViewWorker, e)
		}
	}
	if v.viewWorker != nil {
		v.viewWorker.Layout()
	}

	if v.viewMenu != nil {
		v.viewMenu.Layout()
	}
	log.Layout()
	return
}
func (v *View) increaseWorker(g *gocui.Gui, _ *gocui.View) error {
	v.rely.Increase()
	return nil
}
func (v *View) reduceWorker(g *gocui.Gui, _ *gocui.View) error {
	v.rely.Reduce()
	return nil
}

func (v *View) clickMenu(g *gocui.Gui, view *gocui.View) error {
	x, y := view.Cursor()
	str, _ := view.Word(x, y)
	if str == `` {
		return nil
	}
	return v.doMenu(g, view, y)
}
func (v *View) upMenu(g *gocui.Gui, view *gocui.View) error {
	x, y := view.Cursor()
	if y > 0 {
		view.SetCursor(x, y-1)
	}
	return nil
}
func (v *View) downMenu(g *gocui.Gui, view *gocui.View) error {
	x, y := view.Cursor()
	if y < 3 {
		view.SetCursor(x, y+1)
	}
	return nil
}
func (v *View) enterMenu(g *gocui.Gui, view *gocui.View) error {
	_, y := view.Cursor()
	return v.doMenu(g, view, y)
}
func (v *View) doMenu(g *gocui.Gui, view *gocui.View, y int) error {
	str, e := view.Line(y)
	if e != nil || len(str) < 1 {
		return nil
	}
	switch str[:1] {
	case "m":
		v.toggleMenu(g, nil)
	case "l":
		v.toggleLog(g, nil)
	case "w":
		v.rely.Increase()
	case "s":
		v.rely.Reduce()
	}
	return nil
}
