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

	// start emulation with enabled mmu:
	sys.mmuEnabled = true

	// point mmu to memory:
	mmunit = mmu.MMU{}
	mmunit.Memory = &sys.Memory

	console.WriteConsole("Initializing PDP11 CPU.\n")
	sys.CPU = pdpcpu.New(&mmunit)
	sys.CPU.State = pdpcpu.RUN
	return sys
}

// emulate calls CPU execute as long as cpu is in run state:
func (sys *System) emulate() {
	for sys.CPU.State == pdpcpu.RUN {
		sys.console.WriteConsole(sys.CPU.PrintRegisters() + "\n")
		sys.CPU.Execute()
	}
	sys.console.WriteConsole(
		fmt.Sprintf("CPU status: %v\n ", sys.CPU.State))
}
