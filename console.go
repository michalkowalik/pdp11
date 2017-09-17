package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

/*
group all Status console related functions here
The idea is to run the console in a goroutine

requested functionlity:
	- autoscroll buffer
	- display emulator status messages
	- display prompt at all the time. Prompt symbol: ". " (dot space)
	- requested commands:
	  - START - initialize the cpu / check existing disk images
	  - BOOT <disk> boot from the <disk> -> to be cheked -> shouldn't it be the main
	    program functionality?
	  - HALT - halt the CPU, keep all the values in memory and registers intact
	  - STEP - execute single CPU operation
	  - RESET - reset emulator, run initialization procedure.

	- other elements of the emulator should be able to log information to console
	  using string channel

	- basic line editing functionality should be enabled
*/

// channel to  send console updates to:
var consoleOut chan string

func initConsole(g *gocui.Gui) {
	consoleOut = make(chan string)
	go func() {
		for {
			s := <-consoleOut
			g.Execute(func(g *gocui.Gui) error {
				v, _ := g.View("status")
				fmt.Fprintf(v, "%s\n", s)

				// set cursor one line lower:
				v.MoveCursor(0, 1, true)

				// TODO: needed here?
				return nil
			})
		}
	}()
}

func writeConsole(msg string) error {
	consoleOut <- msg

	// TODO: really needed here?
	return nil
}
