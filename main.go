package main

import (
	"fmt"
	"pdp/system"

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

	fmt.Fprintf(statusView, "Starting PDP-11/70 emulator..\n")
	pdp := system.InitializeSystem(statusView, consoleView)
	pdp.Noop()

	// default return value -> no errors encoutered
	return nil
}

// gocui layout
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	// up -> console
	if v, err := g.SetView("console", 0, 0, maxX-1, maxY-15); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Console"
	}

	// down -> status
	if v, err := g.SetView("status", 0, maxY-14, maxX-1, maxY-1); err != nil {
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
