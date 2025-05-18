package system

import (
	"fmt"
	"go/build"
	"log"
	"path/filepath"
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
	log    *log.Logger

	// console and status output:
	console      console.Console
	terminalView *gocui.View
	regView      *gocui.View
}

var (
	clockCounter uint16
	trapDebug    = true
)

// InitializeSystem initializes the emulated PDP-11/40 hardware
func InitializeSystem(
	c console.Console, terminalView, regView *gocui.View, gui *gocui.Gui, debugMode bool, log *log.Logger) *System {
	sys := new(System)
	sys.console = c
	sys.terminalView = terminalView
	sys.regView = regView
	sys.log = log

	// unibus
	sys.unibus = unibus.New(&sys.psw, gui, &c, debugMode, log)
	sys.unibus.PdpCPU.Reset()

	// mount drive
	fp := filepath.Join(build.Default.GOPATH, "src/pdp11/rk0")
	fmt.Printf("Disk image path: %s\n", fp)
	if err := sys.unibus.Rk01.Attach(0, fp); err != nil {
		panic("Can't mount the drive")
	}

	sys.unibus.Rk01.Reset()
	_ = sys.console.WriteConsole("Initializing PDP11 CPU.\n")

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
			sys.log.Printf("SENDING TRAP %o in the run sys.run : %s\n", t.Vector, t.Msg)
			sys.trap(t)
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

// single cpu step:
func (sys *System) step() {
	// handle interrupts
	if sys.unibus.InterruptQueue[0].Vector > 0 &&
		sys.unibus.InterruptQueue[0].Priority >= sys.psw.Priority() {
		sys.processInterrupt(sys.unibus.InterruptQueue[0])
		for i := 0; i < len(sys.unibus.InterruptQueue)-1; i++ {
			sys.unibus.InterruptQueue[i] = sys.unibus.InterruptQueue[i+1]
		}
		// empty interrupt struct
		sys.unibus.InterruptQueue[len(sys.unibus.InterruptQueue)-1] = interrupts.Interrupt{}
		return
	}

	// execute next CPU instruction
	sys.CPU.Execute()
	clockCounter++
	if clockCounter >= 40000 {
		clockCounter = 0
		sys.unibus.LKS |= 1 << 7
		if sys.unibus.LKS&(1<<6) != 0 {
			sys.unibus.SendInterrupt(6, interrupts.IntCLOCK)
		}
	}
	sys.unibus.Rk01.Step()
	sys.unibus.TermEmulator.Step()
}

// process interrupt in the cpu interrupt queue
//  1. push current PSW and PC to stack
//  2. load PC from interrupt vector
//  3. load PSW from (interrupt vector) + 2
//  4. if previous state mode was User, then set the corresponding bits in PSW
//  5. Return from subprocedure cpu instruction at the end of interrupt procedure
//     makes sure to set the stack and PSW back to where it belongs
func (sys *System) processInterrupt(interrupt interrupts.Interrupt) {
	defer func() {
		t := recover()
		switch t := t.(type) {
		case interrupts.Trap:
			sys.log.Printf("SENDING TRAP %o while processing interrupt: %s\n", t.Vector, t.Msg)
			sys.trap(t)
		case nil:
			break
		default:
			panic(t)
		}

		sys.CPU.Registers[7] = sys.unibus.Mmu.ReadMemoryWord(interrupt.Vector)
		intPSW := sys.unibus.Mmu.ReadMemoryWord(interrupt.Vector + 2)

		if (intPSW & (1 << 14)) != 0 {
			fmt.Printf("ALERT: Fetched Interrupt PSW is in user mode")
		}

		if sys.unibus.Psw.GetPreviousMode() == psw.UserMode {
			intPSW |= (1 << 13) | (1 << 12)
		}
		sys.psw.Set(intPSW)
		sys.CPU.State = unibus.CPURUN
	}()

	// DEBUG: push to interrupt stack
	//if interrupt.Vector != interrupts.IntCLOCK {
	//	fmt.Printf("processing interrupt with the vector 0%o\n", interrupt.Vector)
	//}
	//sys.unibus.InterruptStack.Push(interrupt)

	if interrupt.Vector != interrupts.IntCLOCK {
		sys.log.Printf("processing interrupt with the vector 0%o\n", interrupt.Vector)

	}

	if sys.psw.GetMode() == psw.UserMode {
		fmt.Printf("User mode interrupt\n")
	}

	prev := sys.psw.Get()
	sys.CPU.SwitchMode(psw.KernelMode)
	sys.CPU.Push(prev)
	sys.CPU.Push(sys.CPU.Registers[7])
}

// Trap handles all Trap / abort events.
func (sys *System) trap(trap interrupts.Trap) {
	var prevPSW uint16
	defer func() {
		t := recover()
		switch t := t.(type) {
		case interrupts.Trap:
			fmt.Printf("RED STACK TRAP!")
			sys.unibus.Memory[0] = sys.CPU.Registers[7]
			sys.unibus.Memory[1] = prevPSW
			trap.Vector = 4
			panic("FATAL")
		case nil:
			break
		default:
			panic(t)
		}
	}()

	if trapDebug {
		fmt.Printf("TRAP %o occured: %s\n", trap.Vector, trap.Msg)
	}

	if trap.Vector&1 == 1 {
		panic("Trap called with odd vector number!")
	}

	prevPSW = sys.psw.Get()
	sys.CPU.SwitchMode(psw.KernelMode)
	sys.CPU.Push(prevPSW)
	sys.CPU.Push(sys.CPU.Registers[7])

	sys.CPU.Registers[7] = sys.unibus.ReadIO(unibus.Uint18(trap.Vector))
	sys.unibus.Psw.Set(sys.unibus.ReadIO(unibus.Uint18(trap.Vector) + 2))
	if sys.CPU.IsPrevModeUser() { // user mode
		sys.psw.Set(sys.psw.Get() | (1 << 13) | (1 << 12))
	}
}
