package unibus

import (
	"errors"
	"pdp/disk"
	"pdp/teletype"

	"github.com/jroimartin/gocui"
)

// Interrupt type - used to sygnalize incoming interrupt
// perhaps to be added:
// - delay
// - callback
// - callarg
type Interrupt struct {
	Priority  uint16
	vector    uint16
	cleanFlag bool
}

// const values for memory addresses for attached devices
const (
	VT100Addr = 017772000
)

// Unibus definition
type Unibus struct {

	// Unibus map registers
	UnibusMap [32]int16

	// Channel for interrupt communication
	Interrupts chan Interrupt
}

// attached devices:
var (
	// 1. terminal:
	termEmulator *teletype.Teletype

	// 2. rk01 disk
	rk01 *disk.RK
)

// New initializes and returns the Unibus variable
func New(termView *gocui.View) *Unibus {
	unibus := Unibus{}
	unibus.Interrupts = make(chan Interrupt)

	// initialize attached devices:
	termEmulator = teletype.New(termView)
	return &unibus
}

// map 18 bit unibus address to 22 bit physical via the unibus map (if active)
// TODO: implementation missing
func (u *Unibus) mapUnibusAddress(unibusAddress uint32) uint32 {
	return 0
}

// WriteHello : temp function, just to see if it works at all:
func (u *Unibus) WriteHello() {
	termEmulator.Run()

	termEmulator.TPS = 0x80
	termEmulator.Incoming <- teletype.Instruction{
		Address: 0564,
		Data:    uint16(1 << 6),
		Read:    false}
	termEmulator.Incoming <- teletype.Instruction{
		Address: 0566,
		Data:    0110,
		Read:    false}
}

func (u *Unibus) readIOPage(physicalAddres uint32, byteFlag bool) (uint16, error) {
	switch physicalAddres {
	case VT100Addr:
		// return termEmulator.ReadVT100(byteFlag, physicalAddres)
		return 0, nil
	default:
		return 0, errors.New("Not a UNIBUS Address -> halt / trap?")
	}
}

func (u *Unibus) writeIOPage(physicalAddres uint32, data uint16, byteFlag bool) error {
	switch physicalAddres {
	case VT100Addr:
		// return termEmulator.WriteVT100(byteFlag, physicalAddres, data)
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
