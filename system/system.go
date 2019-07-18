package system

import (
	"fmt"
	"pdp/console"
	"pdp/interrupts"
	"pdp/psw"
	"pdp/unibus"

	"github.com/jroimartin/gocui"
)

// System definition.
type System struct {
	CPU *unibus.CPU
	psw psw.PSW

	// Unibus
	unibus *unibus.Unibus

	// console and status output:
	console      console.Console
	terminalView *gocui.View
	regView      *gocui.View
}

// InitializeSystem initializes the emulated PDP-11/40 hardware
func InitializeSystem(
	c console.Console, terminalView, regView *gocui.View, gui *gocui.Gui) *System {
	sys := new(System)
	sys.console = c
	sys.terminalView = terminalView
	sys.regView = regView

	// unibus
	sys.unibus = unibus.New(&sys.psw, gui, &c)
	sys.unibus.PdpCPU.Reset()

	sys.unibus.WriteHello()
	sys.unibus.WriteHello()

	// mount drive
	// TODO: clean it up!
	sys.unibus.Rk01.Attach(0, "/home/mkowalik/src/pdp/images/rk0.img")
	sys.unibus.Rk01.Reset()

	sys.console.WriteConsole("Initializing PDP11 CPU.\n")
	sys.CPU = sys.unibus.PdpCPU
	sys.CPU.State = unibus.CPURUN
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
		t := recover()
		switch t := t.(type) {
		case interrupts.Trap:
			sys.unibus.Traps <- t
		case nil:
			// ignore
		default:
			panic(t)
		}
	}()

	for {
		sys.step()
	}
}

//  single cpu step:
func (sys *System) step() {
	// handle traps:
	sys.handleTraps()

	// handle interrupts
	if sys.unibus.InterruptQueue[0].Vector > 0 &&
		sys.unibus.InterruptQueue[0].Priority >= sys.psw.Priority() {
		sys.processInterrupt(sys.unibus.InterruptQueue[0])
		for i := 0; i < len(sys.unibus.InterruptQueue)-1; i++ {
			sys.unibus.InterruptQueue[i] = sys.unibus.InterruptQueue[i+1]
		}
		// empty interrupt struct
		sys.unibus.InterruptQueue[len(sys.unibus.InterruptQueue)-1] = interrupts.Interrupt{}

	}

	// register status before execution
	// DEBUG
	sys.console.WriteConsole(sys.CPU.DumpRegisters())

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
	sys.unibus.Rk01.Step()
}

// encapsulate in method
// read from channel in case
// handle trap by possibly calling the already implemented cpu method
func (sys *System) handleTraps() {
	if sys.unibus.ActiveTrap.Vector > 0 {
		// handle trap
		sys.console.WriteConsole(
			fmt.Sprintf(
				"TRAP AT: %d, MSG: %s",
				sys.unibus.ActiveTrap.Vector, sys.unibus.ActiveTrap.Msg))
		sys.CPU.Trap(sys.unibus.ActiveTrap.Vector)
		// set active to nil:
		sys.unibus.ActiveTrap.Vector = 0
		sys.unibus.ActiveTrap.Msg = ""
	}
}

// process interrupt in the cpu interrupt queue
// 1. push current PSW and PC to stack
// 2. load PC from interrupt vector
// 3. load PSW from (interrupt vector) + 2
// 4. if previous state mode was User, then set the corresponding bits in PSW
// 5. Return from subprocedure cpu instruction at the end of interrupt procedure
//    makes sure to set the stack and PSW back to where it belongs
// TODO: wouldn't it make sense to move this method to CPU?
func (sys *System) processInterrupt(interrupt interrupts.Interrupt) {
	prev := sys.psw.Get()
	defer func(prev uint16) {
		t := recover()
		switch t := t.(type) {
		case interrupts.Trap:
			sys.unibus.Traps <- t
		case nil:
			// ignore
		default:
			panic(t)
		}
		sys.CPU.Registers[7] = sys.unibus.Mmu.ReadMemoryWord(interrupt.Vector)
		intPSW := sys.unibus.Mmu.ReadMemoryWord(interrupt.Vector + 2)

		if (prev & (1 << 14)) > 0 {
			intPSW |= (1 << 13) | (1 << 12)
		}
		sys.psw.Set(intPSW)
	}(prev)

	sys.CPU.SwitchMode(psw.KernelMode)
	sys.CPU.Push(prev)
	sys.CPU.Push(sys.CPU.Registers[7])
}
