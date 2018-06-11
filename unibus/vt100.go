package unibus

import (
	"github.com/jroimartin/gocui"
)

// Video terminal emulation. the simpler, the better.
// I'm happy to just see anything on the screen so far.
// not sure if original VT100 could support anything but 7 bit ASCII,
// but that's something we can work on later.

// vt100 (or vt11 - it doesn't matter actually for now) is mapped
// through

// VT100 terminal definition
type VT100 struct {
	// display Program counter
	// 16 bit is a bit of a guess, as I can't really find the info
	// what is the actual size of that counter
	DPC uint16

	// Display Status register
	// 15 -> stop
	// 14, 13, 12, 11 -> mode
	// 10, 9, 8 -> intensity
	// 7 -> pen
	// 6 -> shift
	// 5 -> edge
	// 4 -> italics
	// 3 -> blink
	// 2 -> spare
	// 1, 0 -> line
	DSR uint16

	// gocui view:
	termView *gocui.View

	// horizontal and vertical position of the cursor
	xRegister uint16
	yRegister uint16

	// set to true if debug information required
	debug bool
}

// NewTerm returns pointer to the instance of terminal emulator struct
func NewTerm(termView *gocui.View) *VT100 {
	vt100 := VT100{}
	vt100.termView = termView
	return &vt100
}

// ReadVT100 : read from VT100 memory address.
// data consist of 3 elements: address, data, byte flag.
// upate: it can actually stay here.
func (vt *VT100) ReadVT100(byteFlag bool, physicalAddress uint32) (uint16, error) {
	return 0, nil
}

// WriteVT100 : write to the terminal memory
// last 3 bits in the memory specify the action to execute:
func (vt *VT100) WriteVT100(byteFlag bool, physicalAddress uint32, data uint16) error {
	switch physicalAddress & 06 {
	// DPC
	case 0:
		return nil

	// DSR
	case 2:
		return nil

	// Light pen used by VT11. not really interesting, happy to ignore
	default:
		// TODO: trap needed here.
		return nil
	}
}

// execute command stored in Unibus address for VT100
func (vt *VT100) execute() error {
	return nil
}
