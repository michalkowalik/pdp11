package console

import (
	"fmt"
	"strings"

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

// Console type definition
type Console struct {
	consoleOut  chan string // string channel, to which the console data is sent to
	g           *gocui.Gui  // main gocui GUI object
	v           *gocui.View // gocui view of the control console
	currentLine int         // counter to keep the position of the cursor
}

// New returns a pointer to the new console and runs the initialization procedure:
func New(g *gocui.Gui) *Console {
	c := new(Console)
	c.consoleOut = make(chan string)
	c.g = g
	c.v, _ = g.View("status")
	c.initConsole()
	return c
}

// initConsole initializes the emulator console
func (c *Console) initConsole() {
	go func() {
		for {
			s := <-c.consoleOut
			c.g.Update(func(g *gocui.Gui) error {
				fmt.Fprintf(c.v, "%s", s)

				// TODO: needed here?
				return nil
			})
		}
	}()
}

// WriteConsole displays a string on the console
func (c *Console) WriteConsole(msg string) error {
	for _, line := range strings.Split(msg, "\n") {
		if line != "" {
			c.consoleOut <- line + "\n"
			c.v.MoveCursor(0, 1, true)
			c.currentLine++
		}
	}

	// TODO: really needed here?
	return nil
}
