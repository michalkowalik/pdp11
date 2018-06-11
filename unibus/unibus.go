package unibus

import (
	"errors"
	"pdp/disk"

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
	termEmulator *VT100

	// 2. rk01 disk
	rk01 *disk.RK
)

// New initializes and returns the Unibus variable
func New(termView *gocui.View) *Unibus {
	unibus := Unibus{}
	unibus.Interrupts = make(chan Interrupt)

	// initialize attached devices:
	termEmulator = NewTerm(termView)
	return &unibus
}

// map 18 bit unibus address to 22 bit physical via the unibus map (if active)
// TODO: implementation missing
func (u *Unibus) mapUnibusAddress(unibusAddress uint32) uint32 {
	return 0
}

func (u *Unibus) readIOPage(physicalAddres uint32, byteFlag bool) (uint16, error) {
	switch physicalAddres {
	case VT100Addr:
		return termEmulator.ReadVT100(byteFlag, physicalAddres)
	default:
		return 0, errors.New("Not a UNIBUS Address -> halt / trap?")
	}
}

func (u *Unibus) writeIOPage(physicalAddres uint32, data uint16, byteFlag bool) error {
	switch physicalAddres {
	case VT100Addr:
		return termEmulator.WriteVT100(byteFlag, physicalAddres, data)
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
