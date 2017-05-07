package system

import (
	"fmt"
	"pdp/pdpcpu"

	"github.com/jroimartin/gocui"
)

// System definition.
type System struct {
	Memory [4 * 1024 * 1024]byte
	CPU    *pdpcpu.CPU

	// Unibus map registers
	UnibusMap [32]int16

	// console and status output:
	statusView  *gocui.View
	consoleView *gocui.View
}

// InitializeSystem initializes the emulated PDP-11/44 hardware
func InitializeSystem(statusView, consoleView *gocui.View) *System {
	sys := new(System)
	sys.CPU = new(pdpcpu.CPU)
	sys.statusView = statusView
	sys.consoleView = consoleView
	fmt.Fprintf(statusView, "Initializing PDP11 CPU...\n")
	return sys
}

// Noop is a dummy function just to keep go compiler happy for a while
func (sys *System) Noop() {
	fmt.Fprintf(sys.consoleView, ".. Noop ..\n")
}
