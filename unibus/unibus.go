package unibus

import (
	"errors"
	"pdp/console"
	"pdp/disk"
	"pdp/interrupts"
	"pdp/teletype"

	"github.com/jroimartin/gocui"
)

// const values for memory addresses for attached devices
const (
	VT100Addr = 017772000
)

// Unibus definition
type Unibus struct {

	// Unibus map registers
	UnibusMap [32]int16

	// Channel for interrupt communication
	Interrupts chan interrupts.Interrupt

	// console
	controlConsole *console.Console
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
	return &unibus
}

// map 18 bit unibus address to 22 bit physical via the unibus map (if active)
// TODO: implementation missing
func (u *Unibus) mapUnibusAddress(unibusAddress uint32) uint32 {
	return 0
}

// WriteHello : temp function, just to see if it works at all:
func (u *Unibus) WriteHello() {
	termEmulator.TPS = 0x80
	instructions := []teletype.Instruction{
		teletype.Instruction{
			Address: 0564,
			Data:    uint16(1 << 6),
			Read:    false},
		teletype.Instruction{
			Address: 0566,
			Data:    0111,
			Read:    false},
		teletype.Instruction{
			Address: 0566,
			Data:    0110,
			Read:    false},
	}

	for _, i := range instructions {
		termEmulator.Incoming <- i
	}
}

func (u *Unibus) readIOPage(physicalAddres uint32, byteFlag bool) (uint16, error) {
	switch physicalAddres {
	case VT100Addr:
		return termEmulator.ReadTerm(physicalAddres)
	default:
		return 0, errors.New("Not a UNIBUS Address -> halt / trap?")
	}
}

func (u *Unibus) writeIOPage(physicalAddres uint32, data uint16, byteFlag bool) error {
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
	i := Interrupt{
		Priority: priority,
		vector:   vector}

	// send interrupt:
	go func() { u.Interrupts <- i }()
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
