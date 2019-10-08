package unibus

import (
	"errors"
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

	// Unibus map registers
	// todo: remove if not needed!
	UnibusMap [32]int16

	// LKS - KW11-L Clock status
	LKS uint16

	// Memory management Unit
	Mmu *MMU18Bit

	// Channel for interrupt communication
	Interrupts chan interrupts.Interrupt
	Traps      chan interrupts.Trap

	// console
	controlConsole console.Console

	// terminal emulator
	TermEmulator teletype.Teletype

	// InterruptQueue queue to keep incoming interrupts before processing them
	// TODO: change to array!
	InterruptQueue [8]interrupts.Interrupt

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
	unibus.Interrupts = make(chan interrupts.Interrupt)
	unibus.Traps = make(chan interrupts.Trap)

	// todo: why does it fail on test?
	unibus.controlConsole = *controlConsole
	unibus.psw = psw

	// initialize attached devices:
	unibus.Mmu = NewMMU(psw, &unibus)
	unibus.PdpCPU = NewCPU(unibus.Mmu)

	// TODO: it needs to be modified, in order to allow the GUI!
	unibus.TermEmulator = teletype.NewSimple(unibus.Interrupts) //gui, controlConsole, unibus.Interrupts)
	unibus.TermEmulator.Run()

	unibus.Rk01 = NewRK(&unibus)

	unibus.processInterruptQueue()
	unibus.processTraps()
	return &unibus
}

// save incoming interrupt in a proper place
func (u *Unibus) processInterruptQueue() {
	go func() error {
		for {
			interrupt := <-u.Interrupts

			fmt.Printf("new interrupt\n")

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
	}()
}

// TODO: is there any other way to handle traps actually?
func (u *Unibus) processTraps() {
	go func() error {
		for {
			trap := <-u.Traps
			fmt.Printf("Trap vector: %d, message: \"%s\"\n", trap.Vector, trap.Msg)
			if trap.Vector > 0 {
				u.ActiveTrap = trap
				// TODO: is it actually finished?
				// panic("IT'S A TRAP!!")
			}
		}
	}()
}

// get Register value for address:
func (u *Unibus) getRegisterValue(addr uint32) uint16 {
	reg := (addr & 77) / 2
	return u.PdpCPU.Registers[reg]
}

func (u *Unibus) setRegisterValue(addr uint32, data uint16) {
	reg := (addr & 77) / 2
	u.PdpCPU.Registers[reg] = data
}

// ReadIOPage reads from unibus devices.
func (u *Unibus) ReadIOPage(physicalAddress uint32, byteFlag bool) (uint16, error) {
	switch {
	case physicalAddress == PSWAddr:
		return u.psw.Get(), nil
	case physicalAddress&RegAddr == RegAddr:
		return u.getRegisterValue(physicalAddress), nil
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
// TODO: that signature smells funny. better to resign from that error return type ?
func (u *Unibus) WriteIOPage(physicalAddress uint32, data uint16, byteFlag bool) error {
	switch {
	case physicalAddress == PSWAddr:
		u.psw.Set(data)
		return nil
	case physicalAddress&RegAddr == RegAddr:
		u.setRegisterValue(physicalAddress, data)
		return nil
	case physicalAddress&0777770 == ConsoleAddr:
		u.TermEmulator.WriteTerm(physicalAddress, data)
		return nil
	case physicalAddress == SR0Addr:
		u.Mmu.SR0 = data
		return nil
	case physicalAddress == SR2Addr:
		u.Mmu.SR2 = data
		return nil
	case physicalAddress&0777760 == RK11Addr:
		u.Rk01.write(physicalAddress, data)
		return nil
	case (physicalAddress&0777600 == 0772200) || (physicalAddress&0777600 == 0777600):
		u.Mmu.writePage(physicalAddress, data)
		return nil
	default:
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Write to invalid address %06o", physicalAddress)})
	}
}

// SendInterrupt sends a new interrupts to the receiver
func (u *Unibus) SendInterrupt(priority uint16, vector uint16) {
	i := interrupts.Interrupt{
		Priority: priority,
		Vector:   vector}

	// send interrupt:
	go func() { u.Interrupts <- i }()
}

// SendTrap sends a Trap to CPU the same way the interrupt is sent.
func (u *Unibus) SendTrap(vector uint16, msg string) {
	t := interrupts.Trap{
		Vector: vector,
		Msg:    msg}
	go func() { u.Traps <- t }()

}

// InsertData updates a word with new byte or word data allowing
// for odd addressing
// original        : original value of the data at the address (though, really needed?)
// physicalAddress : address of the value to be changed
// data            : new data to write
// byteFlag        : only access byte, not the complete word
func (u *Unibus) InsertData(
	original uint16, physicalAddres uint32, data uint16, byteFlag bool) error {

	// if odd address:
	if physicalAddres&1 != 0 {

		// trying to access word on odd address
		if !byteFlag {
			return errors.New("Trap needed! -> odd adderss & word set")
		}

	}
	return nil
}
