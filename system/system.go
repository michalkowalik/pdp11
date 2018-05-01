package system

import (
	"fmt"

	"pdp/console"
	"pdp/mmu"
	"pdp/pdpcpu"
	"pdp/unibus"

	"github.com/jroimartin/gocui"
)

// System definition.
type System struct {
	Memory [4 * 1024 * 1024]byte
	CPU    *pdpcpu.CPU

	mmuEnabled bool

	// Unibus
	unibus *unibus.Unibus

	// console and status output:
	console     *console.Console
	consoleView *gocui.View
	regView     *gocui.View
}

// <- make it a part of system type?
// definitely worth rethinking
var mmunit mmu.MMU

// InitializeSystem initializes the emulated PDP-11/44 hardware
func InitializeSystem(console *console.Console, consoleView, regView *gocui.View) *System {
	sys := new(System)
	sys.console = console
	sys.consoleView = consoleView
	sys.regView = regView

	// start emulation with disabled mmu:
	sys.mmuEnabled = false

	// point mmu to memory:
	mmunit = mmu.MMU{}
	mmunit.Memory = &sys.Memory

	console.WriteConsole("Initializing PDP11 CPU...\n")
	sys.CPU = pdpcpu.New(&mmunit)
	sys.CPU.State = pdpcpu.RUN
	return sys
}

// Noop is a dummy function just to keep go compiler happy for a while
func (sys *System) Noop() {
	fmt.Fprintf(sys.consoleView, ".. Noop ..\n")
}
