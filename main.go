package main

import (
	"errors"
	"flag"
	"fmt"
	"pdp/console"
	"pdp/logger"
	"pdp/system"
	"time"

	"log"

	"github.com/jroimartin/gocui"
)

var debugMode *bool

func main() {
	plainMode := flag.Bool("gui", false, "Run program in gui mode")
	debugMode = flag.Bool("debug", false, "Run with CPU debug information")
	flag.Parse()

	if !*plainMode {
		startPdp(nil)
	} else {
		g, err := gocui.NewGui(gocui.OutputNormal)
		if err != nil {
			log.Panicln("Couldn't create gui!")
		}
		defer g.Close()

		g.SetManagerFunc(layout)
		setKeyBindings(g)

		// start emulation
		g.Update(startPdp)
		if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
			log.Panicln(err)
		}
	}

}

func setKeyBindings(g *gocui.Gui) {
	var err error
	if err = g.SetKeybinding("", gocui.KeyF9, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	//
	if err = g.SetKeybinding(
		"",
		gocui.KeyF7,
		gocui.ModNone,
		func(g *gocui.Gui, _ *gocui.View) error {
			_, err := g.SetCurrentView("terminal")
			return err
		}); err != nil {
		log.Panicln(err)
	}

	if err = g.SetKeybinding(
		"",
		gocui.KeyF8,
		gocui.ModNone,
		func(g *gocui.Gui, _ *gocui.View) error {
			_, err := g.SetCurrentView("status")
			return err
		}); err != nil {
		log.Panicln(err)
	}
}

// start pdp11 --> output to either console or status line..
// (btw -- will it provide the buffer to show the most recent lines?)
func startPdp(g *gocui.Gui) error {
	var (
		c            console.Console
		statusView   *gocui.View
		terminalView *gocui.View
		regView      *gocui.View
		err          error
	)

	if g != nil {
		statusView, err = g.View("status")
		if err != nil {
			return err
		}
		statusView.Clear()

		terminalView, err = g.View("terminal")
		if err != nil {
			return err
		}
		terminalView.Clear()

		regView, err = g.View("registers")
		if err != nil {
			return err
		}
		terminalView.Clear()
		c = console.NewGui(g)

		if _, err := g.SetCurrentView("status"); err != nil {
			log.Panic(err)
		}
		g.Cursor = true
		g.Highlight = true
	} else {
		c = console.NewSimple()
	}

	log := logger.New("pdp11.log")

	c.WriteConsole("Starting PDP-11/40 emulator.")
	pdp := system.InitializeSystem(c, terminalView, regView, g, *debugMode, log)

	// update registers:
	if g != nil {
		updateRegisters(pdp, g)
	}
	pdp.Boot()
	log.Printf("Booting pdp..")

	// default return value -> no errors encountered
	return nil
}

// update registers display
// has to be run in go routine -> gocui allows updating the view only through Execute function
func updateRegisters(pdp *system.System, g *gocui.Gui) {
	ticker := time.NewTicker(time.Second * 1)

	go func() {
		i := 0
		for range ticker.C {

			g.Update(func(g *gocui.Gui) error {
				v, err := g.View("registers")
				if err != nil {
					return err
				}
				v.Clear()
				fmt.Fprint(v, pdp.CPU.DumpRegisters())
				return nil
			})
			i++
		}
	}()
}

// gocui layout
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	// up -> console
	if v, err := g.SetView("terminal", 0, 0, maxX-1, maxY-18); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "|Terminal| [F7]"
		v.Editable = true
		v.Autoscroll = true
	}

	// middle -> register values
	if v, err := g.SetView("registers", 0, maxY-17, maxX-1, maxY-14); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "|Registers|"
	}
	// down -> status
	if v, err := g.SetView("status", 0, maxY-13, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "|System Control Console| [F8]"
		v.Editable = true
		v.Autoscroll = true
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
