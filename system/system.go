package system

import (
	"fmt"

	"pdp/mmu"
	"pdp/pdpcpu"

	"github.com/jroimartin/gocui"
)

// System definition.
type System struct {
	Memory [4 * 1024 * 1024]byte
	CPU    *pdpcpu.CPU

	mmuEnabled bool

	// Unibus map registers
	UnibusMap [32]int16

	// console and status output:
	statusView  *gocui.View
	consoleView *gocui.View
	regView     *gocui.View
}

// <- make it a part of system type?
// definitely worth rethinking
var mmunit mmu.MMU

// InitializeSystem initializes the emulated PDP-11/44 hardware
func InitializeSystem(statusView, consoleView, regView *gocui.View) *System {
	sys := new(System)
	sys.statusView = statusView
	sys.consoleView = consoleView
	sys.regView = regView

	// start emulation with disabled mmu:
	sys.mmuEnabled = false

	// point mmu to memory:
	mmunit = mmu.MMU{}
	mmunit.Memory = &sys.Memory

	fmt.Fprintf(statusView, "Initializing PDP11 CPU...\n")
	sys.CPU = pdpcpu.New(&mmunit)

	return sys
}

// Noop is a dummy function just to keep go compiler happy for a while
func (sys *System) Noop() {
	fmt.Fprintf(sys.consoleView, ".. Noop ..\n")
}
