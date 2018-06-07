package unibus

import (
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

// Unibus definition
type Unibus struct {

	// Unibus map registers
	UnibusMap [32]int16

	// Channel for interrupt communication
	Interrupts chan Interrupt

	// Unibus memory map. rough and uncomplete for now.
	// the variadic parameters are required due to some
	// of the IOPage function requireding 4 parameters
	// TODO: 4'th parameter being...
	memoryMap map[uint32](func(bool, ...uint32) (uint16, error))
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

	// initialize memory map:
	unibus.memoryMap = make(map[uint32](func(bool, ...uint32) (uint16, error)))
	unibus.memoryMap[017772000] = unibus.accessVT100

	return &unibus
}

// map 18 bit unibus address to 22 bit physical via the unibus map (if active)
//  TODO: implementation missing
func (u *Unibus) mapUnibusAddress(unibusAddress uint32) uint32 {
	return 0
}

// access IO Page - write or read.
// TODO: implementation.
func (u *Unibus) accessIOPage(physicalAddres uint32, data uint16, byteFlag bool) error {
	if val, ok := u.memoryMap[physicalAddres]; ok {
		val(byteFlag, physicalAddres, uint32(data))
	}
	return nil
}

// TODO: -> separate accessIOPage into read and write functions:

func (u *Unibus) readIOPage(physicalAddres uint32, byteFlag bool) (uint16, error) {
	if val, ok := u.memoryMap[physicalAddres]; ok {
		// not sure about that 0 -> perhaps that should be separeted too!
		return val(byteFlag, physicalAddres, 0)
	}
	return 0, nil

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
