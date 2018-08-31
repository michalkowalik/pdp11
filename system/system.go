package system

import (
	"pdp/console"
	"pdp/interrupts"
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

// InitializeSystem initializes the emulated PDP-11/40 hardware
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

// Run system
func (sys *System) Run() {
	for {
		sys.run()
	}
}

// actually run the system
func (sys *System) run() {
	defer func() {
		// recover from trap...
	}()

	for {
		sys.step()
	}
}

//  single cpu step:
func (sys *System) step() {
	// handle interrupts
	sys.processInterruptQueue()

	// execute next CPU instruction
	sys.CPU.Execute()
	sys.CPU.ClockCounter++
	if sys.CPU.ClockCounter >= 40000 {
		sys.CPU.ClockCounter = 0
		sys.unibus.LKS |= (1 << 7)
		if sys.unibus.LKS&(1<<6) != 0 {
			sys.unibus.SendInterrupt(6, interrupts.INTClock)
		}
	}
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

// process interrupt in the cpu interrup queue
func (sys *System) processInterrupt(interrupt interrupts.Interrupt) {

}
