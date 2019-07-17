package console

import (
	"os"
	"strings"
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

// Simple console type definition
type Simple struct {
	consoleOut  chan string // string channel, to which the console data is sent to
	currentLine int         // counter to keep the position of the cursor
}

// NewSimple returns a pointer to the new console and runs the initialization procedure:
func NewSimple() *Simple {
	c := new(Simple)
	c.consoleOut = make(chan string)
	c.initSimple()
	return c
}

// initSimple initializes the emulator console
// TODO: Do I want a select mutex here?
func (c *Simple) initSimple() {
	go func() {
		for {
			s := <-c.consoleOut
			os.Stdout.Write([]byte(s))
		}
	}()
}

// WriteConsole displays a string on the console
func (c *Simple) WriteConsole(msg string) error {
	for _, line := range strings.Split(msg, "\n") {
		if line != "" {
			c.consoleOut <- line + "\n"
			c.currentLine++
		}
	}
	return nil
}
