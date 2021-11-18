package get

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

const (
	ViewClose = `close`
	ViewConf  = `conf`
	ViewLog   = `log`
)

type GUI struct {
	conf *Configure
	log  string
}

func NewGUI(conf *Configure) *GUI {
	return &GUI{
		conf: conf,
	}
}
func (gui *GUI) Serve() (e error) {
	g, e := gocui.NewGui(gocui.OutputNormal)
	if e != nil {
		return
	}
	defer g.Close()
	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen
	g.Mouse = true

	g.SetManagerFunc(gui.layout)
	if e = g.SetKeybinding(``, gocui.KeyCtrlC, gocui.ModNone, gui.quit); e != nil {
		return
	} else if e = g.SetKeybinding(``, gocui.MouseLeft, gocui.ModNone, gui.current); e != nil {
		return
	} else if e = g.SetKeybinding(``, gocui.MouseWheelUp, gocui.ModNone, gui.wheelUp); e != nil {
		return
	} else if e = g.SetKeybinding(``, gocui.MouseWheelDown, gocui.ModNone, gui.wheelDown); e != nil {
		return
	}
	e = g.MainLoop()
	if e == gocui.ErrQuit {
		e = nil
	}
	return
}

func (gui *GUI) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
func (gui *GUI) current(g *gocui.Gui, v *gocui.View) (e error) {
	name := v.Name()
	if _, e = g.SetCurrentView(name); e != nil {
		return
	}
	return
}
func (gui *GUI) wheelUp(g *gocui.Gui, v *gocui.View) (e error) {
	x, y := v.Origin()
	if y > 0 {
		e = v.SetOrigin(x, y-1)
	}

	return
}
func (gui *GUI) wheelDown(g *gocui.Gui, v *gocui.View) (e error) {
	count := len(v.ViewBufferLines())
	_, sy := v.Size()

	x, y := v.Origin()
	y++
	if y+sy >= count {
		return
	}
	e = v.SetOrigin(x, y+1)
	if e != nil {
		return
	}
	return
}

func (gui *GUI) layout(g *gocui.Gui) (e error) {
	e = gui.layoutConfigure(g)
	if e != nil {
		return
	}
	e = gui.layoutLog(g)
	if e != nil {
		return
	}
	return
}

func (gui *GUI) layoutConfigure(g *gocui.Gui) (e error) {
	maxX, _ := g.Size()
	v, e := g.SetView(ViewConf,
		0, 0,
		maxX-1, 1+6,
	)
	if e != nil {
		if e != gocui.ErrUnknownView {
			return
		}
		e = nil
		return
	}
	v.Title = ViewConf
	v.Wrap = true
	v.Clear()

	conf := gui.conf
	fmt.Fprintf(v, "      URL: %s\n", conf.URL+conf.URL+conf.URL)
	fmt.Fprintf(v, "   Output: %s\n", conf.Output)
	fmt.Fprintf(v, "    Proxy: %s\n", conf.Proxy)
	fmt.Fprintf(v, "UserAgent: %s\n", conf.UserAgent)
	fmt.Fprintf(v, "     Head: %v\n", conf.Head)
	fmt.Fprintf(v, "   Worker: %v   Block: %v\n", conf.Worker, conf.Block)

	return nil
}
func (gui *GUI) layoutLog(g *gocui.Gui) (e error) {
	maxX, maxY := g.Size()
	v, e := g.SetView(ViewLog,
		0, maxY-5,
		maxX-1, maxY-1,
	)
	if e != nil {
		if e != gocui.ErrUnknownView {
			return
		}
		e = nil
		return
	}
	v.Title = ViewLog
	v.Wrap = true
	v.Clear()
	fmt.Fprintf(v, gui.log)
	return nil
}
