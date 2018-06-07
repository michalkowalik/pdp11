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
	DPC uint32

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

// termporary, probably broken. should be a part of VT100:
// data consist of 3 elements: address, data, byte flag.
// upate: it can actually stay here.
func (u *Unibus) accessVT100(byteFlag bool, data ...uint32) (uint16, error) {
	return 0, nil
}
