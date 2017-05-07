package main

import (
	"fmt"
	"pdp/system"
	"time"

	"log"

	"github.com/jroimartin/gocui"
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln("Couldn't create gui!")
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	// start emulation
	g.Execute(startPdp)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

// start pdp11 --> output to either console or status line..
// (btw -- will it provide the buffer to show the most recent lines?)
func startPdp(g *gocui.Gui) error {
	statusView, err := g.View("status")
	if err != nil {
		return err
	}
	statusView.Clear()

	consoleView, err := g.View("console")
	if err != nil {
		return err
	}
	consoleView.Clear()

	regView, err := g.View("registers")
	if err != nil {
		return err
	}
	consoleView.Clear()

	fmt.Fprintf(statusView, "Starting PDP-11/70 emulator..\n")
	pdp := system.InitializeSystem(statusView, consoleView, regView)

	// update registers:
	updateRegisters(pdp, g)
	pdp.Noop()

	// default return value -> no errors encoutered
	return nil
}

// update registers display
// has to be run in go routine -> gocui allows updating the view only through Execute function
func updateRegisters(pdp *system.System, g *gocui.Gui) {
	ticker := time.NewTicker(time.Second * 1)

	go func() {
		i := 0
		for range ticker.C {

			g.Execute(func(g *gocui.Gui) error {
				v, err := g.View("registers")
				if err != nil {
					return err
				}
				v.Clear()
				pdp.CPU.DumpRegisters(v)
				fmt.Fprintf(v, " <t : 0x%x>", i)
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
	if v, err := g.SetView("console", 0, 0, maxX-1, maxY-18); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Console"
	}

	// middle -> register values
	if v, err := g.SetView("registers", 0, maxY-17, maxX-1, maxY-14); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Registers"
	}
	// down -> status
	if v, err := g.SetView("status", 0, maxY-13, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Status"
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
