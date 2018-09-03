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

	// system stack pointers: kernel, super, illegal, user
	// super won't be needed for pdp11/40:
	KernelStackPointer uint16
	UserStackPointer   uint16

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

// SwitchMode switches the kernel / user mode:
// 0 for user, 3 for kernel, everything else is a mistake.
// values are as they are used in the PSW
func (sys *System) SwitchMode(m uint16) {
	sys.psw.SwitchMode(m)

	// save processor stack pointers:
	if m > 0 {
		sys.KernelStackPointer = sys.CPU.Registers[6]
	} else {
		sys.UserStackPointer = sys.CPU.Registers[6]
	}

	// set processor stack:
	if m > 0 {
		sys.CPU.Registers[6] = sys.KernelStackPointer
	} else {
		sys.CPU.Registers[6] = sys.UserStackPointer
	}
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
	if len(sys.unibus.InterruptQueue) > 0 &&
		sys.unibus.InterruptQueue[0].Priority >= sys.psw.Priority() {
		sys.processInterrupt(sys.unibus.InterruptQueue[0])
		for i := 0; i < len(sys.unibus.InterruptQueue); i++ {
			sys.unibus.InterruptQueue[i] = sys.unibus.InterruptQueue[i+1]
		}
		// empty interrupt struct
		sys.unibus.InterruptQueue[len(sys.unibus.InterruptQueue)-1] = interrupts.Interrupt{}

	}

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

// process interrupt in the cpu interrup queue
func (sys *System) processInterrupt(interrupt interrupts.Interrupt) bool {
	defer func() {
		t := recover()
		switch t := t.(type) {
		case nil:
		default:
			panic(t)

		}
		// TODO: make sure it's not stuck in waiting state!
	}()
	// prev := sys.psw.Get()

	return true
}
