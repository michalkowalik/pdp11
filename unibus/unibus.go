package unibus

import (
	"fmt"
	"pdp/console"
	"pdp/interrupts"
	"pdp/psw"
	"pdp/teletype"

	"github.com/jroimartin/gocui"
)

// Unibus address mappings for attached devices.
const (
	LKSAddr     = 0777546
	ConsoleAddr = 0777560
	RK11Addr    = 0777400
	PSWAddr     = 0777776
	SR0Addr     = 0777572
	SR2Addr     = 0777576
	RegAddr     = 0777700
)

// Unibus definition
type Unibus struct {

	// LKS - KW11-L Clock status
	LKS uint16

	// Memory management Unit
	Mmu *MMU18Bit

	// console
	controlConsole console.Console

	// terminal emulator
	TermEmulator teletype.Teletype

	// InterruptQueue queue to keep incoming interrupts before processing them
	InterruptQueue interrupts.InterruptQueue

	// ActiveTrap keeps the active trap in case the trap is being throw
	// or nil otheriwse
	ActiveTrap interrupts.Trap

	psw *psw.PSW

	PdpCPU *CPU

	Rk01 *RK11
}

// New initializes and returns the Unibus variable
func New(psw *psw.PSW, gui *gocui.Gui, controlConsole *console.Console) *Unibus {
	unibus := Unibus{}

	unibus.controlConsole = *controlConsole
	unibus.psw = psw

	// initialize attached devices:
	unibus.Mmu = NewMMU(psw, &unibus)
	unibus.PdpCPU = NewCPU(unibus.Mmu)

	// TODO: it needs to be modified, in order to allow the GUI!
	unibus.TermEmulator = teletype.NewSimple(&unibus.InterruptQueue)
	unibus.TermEmulator.Run()

	unibus.Rk01 = NewRK(&unibus)
	return &unibus
}

//SendInterrupt : save incoming interrupt in interrupt table
func (u *Unibus) SendInterrupt(priority uint16, vector uint16) {
	fmt.Printf("IRQ: %06o\n", vector)

	interrupt := interrupts.Interrupt{
		Priority: priority,
		Vector:   vector}

	if interrupt.Vector&1 == 1 {
		panic("Interrupt with Odd vector number")
	}

	var i int
	for ; i < len(u.InterruptQueue); i++ {
		if u.InterruptQueue[i].Vector == 0 ||
			u.InterruptQueue[i].Priority < interrupt.Priority {
			break
		}
	}

	for ; i < len(u.InterruptQueue); i++ {
		if u.InterruptQueue[i].Vector == 0 ||
			u.InterruptQueue[i].Vector >= interrupt.Vector {
			break
		}
	}

	if i == len(u.InterruptQueue) {
		panic("Interrupt table full")
	}

	for j := len(u.InterruptQueue) - 1; j > i; j-- {
		u.InterruptQueue[j] = u.InterruptQueue[j-1]
	}
	u.InterruptQueue[i] = interrupt
}

// get Register value for address:
func (u *Unibus) getRegisterValue(addr uint32) uint16 {
	return u.PdpCPU.Registers[addr&07]
}

func (u *Unibus) setRegisterValue(addr uint32, data uint16) {
	u.PdpCPU.Registers[addr&07] = data
}

// ReadIOPage reads from unibus devices.
func (u *Unibus) ReadIOPage(physicalAddress uint32, byteFlag bool) (uint16, error) {
	switch {
	case physicalAddress == PSWAddr:
		return u.psw.Get(), nil
	case physicalAddress&RegAddr == RegAddr:
		return u.getRegisterValue(physicalAddress), nil
	// physical front console. Magic number that seems to do the job:
	case physicalAddress == 0777570:
		return 0173030, nil
	case physicalAddress == LKSAddr:
		return u.LKS, nil
	case physicalAddress&0777770 == ConsoleAddr:
		return u.TermEmulator.ReadTerm(physicalAddress)
	case physicalAddress == SR0Addr:
		return u.Mmu.SR0, nil
	case physicalAddress == SR2Addr:
		return u.Mmu.SR2, nil
	case physicalAddress&0777760 == RK11Addr:
		v := u.Rk01.read(physicalAddress)
		return v, nil
	case (physicalAddress&0777600 == 0772200) || (physicalAddress&0777600 == 0777600):
		return u.Mmu.readPage(physicalAddress), nil
	default:
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Read from invalid address %06o", physicalAddress)})
	}
}

// WriteIOPage writes to the unibus connected device
func (u *Unibus) WriteIOPage(physicalAddress uint32, data uint16, byteFlag bool) {
	switch {
	case physicalAddress == PSWAddr:
		// also : switch mode!
		u.PdpCPU.SwitchMode(data >> 14)
		// also: set flags:
		u.psw.Set(data)
	case physicalAddress&RegAddr == RegAddr:
		u.setRegisterValue(physicalAddress, data)
	case physicalAddress == LKSAddr:
		u.LKS = data
	case physicalAddress&0777770 == ConsoleAddr:
		u.TermEmulator.WriteTerm(physicalAddress, data)
	case physicalAddress == SR0Addr:
		u.Mmu.SR0 = data
	case physicalAddress == SR2Addr:
		u.Mmu.SR2 = data
	case physicalAddress&0777760 == RK11Addr:
		u.Rk01.write(physicalAddress, data)
	case (physicalAddress&0777600 == 0772200) || (physicalAddress&0777600 == 0777600):
		u.Mmu.writePage(physicalAddress, data)
	default:
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Write to invalid address %06o", physicalAddress)})
	}
}
