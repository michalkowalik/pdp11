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
	PSWVirtAddr = 0177776
	SR0Addr     = 0777572
	SR2Addr     = 0777576
	RegAddr     = 0777700
	MEMSIZE     = 0760000 // useful memory. everything above 248K is unibus reserved
)

// Unibus definition
type Unibus struct {
	Memory [MEMSIZE >> 1]uint16

	// LKS - KW11-L Clock status
	LKS uint16

	// Memory management Unit
	Mmu MMU

	// console
	controlConsole console.Console

	// terminal emulator
	TermEmulator teletype.Teletype

	// InterruptQueue queue to keep incoming interrupts before processing them
	InterruptQueue interrupts.InterruptQueue

	// ActiveTrap keeps the active trap in case the trap is being throw
	// or nil otherwise
	ActiveTrap interrupts.Trap

	Psw *psw.PSW

	PdpCPU *CPU

	Rk01 *RK11
}

// New initializes and returns the Unibus variable
func New(psw *psw.PSW, gui *gocui.Gui, controlConsole *console.Console, debugMode bool) *Unibus {
	unibus := Unibus{}

	unibus.controlConsole = *controlConsole
	unibus.Psw = psw

	// initialize attached devices:
	unibus.Mmu = NewMMU18(&unibus)
	unibus.PdpCPU = NewCPU(unibus.Mmu, &unibus, debugMode)

	// TODO: it needs to be modified, in order to allow the GUI!
	unibus.TermEmulator = teletype.NewSimple(&unibus.InterruptQueue)
	if err := unibus.TermEmulator.Run(); err != nil {
		panic("Can't initialize terminal emulator")
	}
	unibus.Rk01 = NewRK(&unibus)
	return &unibus
}

// SendInterrupt : save incoming interrupt in interrupt table
func (u *Unibus) SendInterrupt(priority uint16, vector uint16) {
	u.InterruptQueue.SendInterrupt(priority, vector)
}

// get Register value for address:
func (u *Unibus) getRegisterValue(addr Uint18) uint16 {
	return u.PdpCPU.Registers[addr&07]
}

func (u *Unibus) setRegisterValue(addr uint32, data uint16) {
	u.PdpCPU.Registers[addr&07] = data
}

// ReadIO reads from unibus devices.
func (u *Unibus) ReadIO(physicalAddress Uint18) uint16 {
	switch {
	case physicalAddress&1 == 1:
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Read from the odd address %06o", physicalAddress)})
	case physicalAddress < MEMSIZE:
		return u.Memory[physicalAddress>>1]
	case physicalAddress == PSWAddr:
		return u.Psw.Get()
	case physicalAddress&RegAddr == RegAddr:
		return u.getRegisterValue(physicalAddress)
	// physical front console. Magic number that seems to do the job:
	case physicalAddress == 0777570:
		return 0173030
	case physicalAddress == LKSAddr:
		return u.LKS
	case physicalAddress&0777770 == ConsoleAddr:
		return u.TermEmulator.ReadTerm(uint32(physicalAddress))
	case physicalAddress == SR0Addr:
		return u.Mmu.GetSR0()
	case physicalAddress == SR2Addr:
		return u.Mmu.GetSR2()
	case physicalAddress&0777760 == RK11Addr:
		v := u.Rk01.read(physicalAddress)
		return v
	case (physicalAddress&0777600 == 0772200) || (physicalAddress&0777600 == 0777600):
		return u.Mmu.Read16(physicalAddress)
	default:
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Read from invalid address %06o", physicalAddress)})
	}
}

//TODO: Finish
func (u *Unibus) ReadIOByte(physicalAddress Uint18) uint16 {
	val := u.ReadIO(physicalAddress & ^Uint18(1))
	if physicalAddress&1 != 0 {
		return val >> 8
	}
	return val & 0xFF
}

// WriteIO writes to the unibus connected device
func (u *Unibus) WriteIO(physicalAddress Uint18, data uint16) {
	switch {
	case physicalAddress&1 == 1:
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Write the odd address %06o", physicalAddress)})
	case physicalAddress < MEMSIZE:
		u.Memory[physicalAddress>>1] = data
	case physicalAddress == PSWAddr:
		// also : switch mode!
		u.PdpCPU.SwitchMode(data >> 14)
		// also: set flags:
		u.Psw.Set(data)
	case physicalAddress&RegAddr == RegAddr:
		u.setRegisterValue(uint32(physicalAddress), data)
	case physicalAddress == LKSAddr:
		u.LKS = data
	case physicalAddress&0777770 == ConsoleAddr:
		_ = u.TermEmulator.WriteTerm(uint32(physicalAddress), data)
	case physicalAddress == SR0Addr:
		u.Mmu.SetSR0(data)
	case physicalAddress == SR2Addr:
		u.Mmu.SetSR2(data)
	case physicalAddress&0777760 == RK11Addr:
		u.Rk01.write(uint32(physicalAddress), data)
	case (physicalAddress&0777600 == 0772200) || (physicalAddress&0777600 == 0777600):
		u.Mmu.Write16(physicalAddress, data)
	default:
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Write to invalid address %06o", physicalAddress)})
	}
}

// TODO: Finish
func (u *Unibus) WriteIOByte(physicalAddress Uint18, data uint16) {
	// nothing to see here yet
}
