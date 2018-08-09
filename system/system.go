package system

import (
	"fmt"
	"pdp/console"
	"pdp/mmu"
	"pdp/pdpcpu"
	"pdp/psw"
	"pdp/unibus"

	"github.com/jroimartin/gocui"
)

// System definition.
type System struct {
	CPU *pdpcpu.CPU

	psw psw.PSW

	// Unibus
	unibus *unibus.Unibus

	// console and status output:
	console      *console.Console
	terminalView *gocui.View
	regView      *gocui.View
}

// <- make it a part of system type?
// definitely worth rethinking
var mmunit *mmu.MMU18Bit

// InitializeSystem initializes the emulated PDP-11/44 hardware
func InitializeSystem(
	console *console.Console, terminalView, regView *gocui.View, gui *gocui.Gui) *System {
	sys := new(System)
	sys.console = console
	sys.terminalView = terminalView
	sys.regView = regView

	// unibus
	sys.unibus = unibus.New(gui, console)

	// point mmu to memory:
	mmunit = mmu.New(&sys.psw, sys.unibus)

	sys.unibus.WriteHello()
	sys.unibus.WriteHello()

	console.WriteConsole("Initializing PDP11 CPU.\n")
	sys.CPU = pdpcpu.New(mmunit)
	sys.CPU.State = pdpcpu.RUN
	return sys
}

// emulate calls CPU execute as long as cpu is in run state:
// TODO: Probably obsolete. Consider removal
func (sys *System) emulate() {
	for sys.CPU.State == pdpcpu.RUN {
		sys.console.WriteConsole(sys.CPU.PrintRegisters() + "\n")
		sys.CPU.Execute()
	}
	sys.console.WriteConsole(
		fmt.Sprintf("CPU status: %v\n ", sys.CPU.State))
}

// loop keeps the emulation running.
// checks the interrupt queue and lets CPU run
//
// * why does sleep make the memory to explode and blocks everything?
// * how do I actually handle the wait state in a secure manner?

func (sys *System) loop() {
	for {
		for step := 0; step < 4000; step++ {
			// check interrupts
			sys.processInterruptQueue()

			// check traps

			// run cpu instruction
			sys.CPU.Execute()
		}
		sys.console.WriteConsole(sys.CPU.PrintRegisters() + "\n")
		if sys.CPU.State == pdpcpu.WAIT {
			sys.console.WriteConsole("CPU in WAIT state \n")
			break
		}
	}
	sys.console.WriteConsole("out of for loop")
}

// check for incoming interrupt, insert it into CPU interrupt queue.
// multiplexing with `select` to avoid blocking the programme.
func (sys *System) processInterruptQueue() {
	select {
	case interrupt := <-sys.unibus.Interrupts:
		sys.CPU.InterruptQueue = append(sys.CPU.InterruptQueue, interrupt)
		return
	default:
	}
}
