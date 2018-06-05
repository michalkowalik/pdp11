package unibus

// Video terminal emulation. the simpler, the better.
// I'm happy to just see anything on the screen so far.
// not sure if original VT100 could support anything but 7 bit ASCII,
// but that's something we can work on later.

// VT100 terminal definition
type VT100 struct {

	// horizontal and vertical position of the cursor
	xRegister uint16
	yRegister uint16

	// set to true if debug information required
	debug bool
}
