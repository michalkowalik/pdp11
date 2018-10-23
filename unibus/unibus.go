package unibus

import (
	"errors"
	"fmt"
	"pdp/console"
	"pdp/disk"
	"pdp/interrupts"
	"pdp/teletype"

	"github.com/jroimartin/gocui"
)

// Unibus address mappings for attached devices.
const (
	LKSAddr     = 0777546
	ConsoleAddr = 0777560
	RK11Addr    = 0777400
)

// Unibus definition
type Unibus struct {

	// Unibus map registers
	// todo: remove if not needed!
	UnibusMap [32]int16

	// LKS - KW11-L Clock status
	LKS uint16

	// Channel for interrupt communication
	Interrupts chan interrupts.Interrupt
	Traps      chan interrupts.Trap

	// console
	controlConsole *console.Console

	// InterruptQueue queue to keep incoming interrupts before processing them
	// TODO: change to array!
	InterruptQueue [8]interrupts.Interrupt

	// ActiveTrap keeps the active trap in case the trap is being throw
	// or nil otheriwse
	ActiveTrap interrupts.Trap
}

// attached devices:
var (
	// 0. CPU

	// 1. MMU

	// 2. terminal:
	termEmulator *teletype.Teletype

	// 3. rk01 disk
	rk01 *disk.RK
)

// New initializes and returns the Unibus variable
func New(gui *gocui.Gui, controlConsole *console.Console) *Unibus {
	unibus := Unibus{}
	unibus.Interrupts = make(chan interrupts.Interrupt)
	unibus.Traps = make(chan interrupts.Trap)
	unibus.controlConsole = controlConsole

	// initialize attached devices:
	termEmulator = teletype.New(gui, controlConsole, unibus.Interrupts)
	termEmulator.Run()
	unibus.processInterruptQueue()
	unibus.processTraps()
	return &unibus
}

// save incoming interrupt in a proper place
func (u *Unibus) processInterruptQueue() {
	go func() error {
		for {
			interrupt := <-u.Interrupts

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
				// panic("IT'S A TRAP!!")
			}
		}
	}()
}

// WriteHello : temp function, just to see if it works at all:
func (u *Unibus) WriteHello() {
	helloStr := "0_1.2_3.4_5.6_7.8_9.A_B.C_D.E_F\n"
	for _, c := range helloStr {
		termEmulator.Incoming <- teletype.Instruction{
			Address: 0566,
			Data:    uint16(c),
			Read:    false}
	}
}

// ReadIOPage reads from unibus devices.
func (u *Unibus) ReadIOPage(physicalAddress uint32, byteFlag bool) (uint16, error) {
	switch {
	case physicalAddress&0777770 == ConsoleAddr:
		return termEmulator.ReadTerm(physicalAddress)
	case physicalAddress&0777760 == RK11Addr:
		// don't do any anything yet!
		return 0, nil
	default:
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Read from invalid address %06o", physicalAddress)})
	}
}

// WriteIOPage writes to the unibus connected device
func (u *Unibus) WriteIOPage(physicalAddress uint32, data uint16, byteFlag bool) error {
	switch {
	case physicalAddress&0777770 == ConsoleAddr:
		termEmulator.Incoming <- teletype.Instruction{
			Address: physicalAddress,
			Data:    data,
			Read:    false}
		return nil
	case physicalAddress&0777760 == RK11Addr:
		// don't do anything yet!
	default:
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Write to invalid address %06o", physicalAddress)})
	}
	return nil
}

// SendInterrupt sends a new interrupts to the receiver
// TODO: implementation!
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
