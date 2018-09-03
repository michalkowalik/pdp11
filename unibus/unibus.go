package unibus

import (
	"errors"
	"pdp/console"
	"pdp/disk"
	"pdp/interrupts"
	"pdp/teletype"

	"github.com/jroimartin/gocui"
)

// Unibus address mappings for attached devices.
const (
	VT100Addr   = 017772000
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

	// console
	controlConsole *console.Console

	// InterruptQueue queue to keep incoming interrupts before processing them
	// TODO: change to array!
	InterruptQueue [8]interrupts.Interrupt
}

// attached devices:
var (
	// 1. terminal:
	termEmulator *teletype.Teletype

	// 2. rk01 disk
	rk01 *disk.RK
)

// New initializes and returns the Unibus variable
func New(gui *gocui.Gui, controlConsole *console.Console) *Unibus {
	unibus := Unibus{}
	unibus.Interrupts = make(chan interrupts.Interrupt)
	unibus.controlConsole = controlConsole

	// initialize attached devices:
	termEmulator = teletype.New(gui, controlConsole, unibus.Interrupts)
	termEmulator.Run()
	unibus.processInterruptQueue()
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

// map 18 bit unibus address to 22 bit physical via the unibus map (if active)
// TODO: implementation missing
func (u *Unibus) mapUnibusAddress(unibusAddress uint32) uint32 {
	return 0
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
func (u *Unibus) ReadIOPage(physicalAddres uint32, byteFlag bool) (uint16, error) {
	switch physicalAddres {
	case VT100Addr:
		return termEmulator.ReadTerm(physicalAddres)
	default:
		return 0, errors.New("Not a UNIBUS Address -> halt / trap?")
	}
}

// WriteIOPage writes to the unibus connected device
func (u *Unibus) WriteIOPage(physicalAddres uint32, data uint16, byteFlag bool) error {
	switch physicalAddres {
	case VT100Addr:
		termEmulator.Incoming <- teletype.Instruction{
			Address: physicalAddres,
			Data:    data,
			Read:    false}
		return nil
	default:
		return errors.New("Not a unibus address -> trap / halt perhaps?")
	}
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
func (u *Unibus) SendTrap(vector uint16) {
	// nothing to see here yet!
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

// Error wrapper : take error, send trap, return error
func (u *Unibus) Error(err error, trapVector uint16) error {
	u.SendTrap(trapVector)
	return err
}
