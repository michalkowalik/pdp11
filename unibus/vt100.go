package unibus

// Video terminal emulation. the simpler, the better.
// I'm happy to just see anything on the screen so far.
// not sure if original VT100 could support anything but 7 bit ASCII,
// but that's something we can work on later.

// vt100 (or vt11 - it doesn't matter actually for now) is mapped
// through

// VT100 terminal definition
type VT100 struct {

	// horizontal and vertical position of the cursor
	xRegister uint16
	yRegister uint16

	// set to true if debug information required
	debug bool
}

// termporary, probably broken. should be a part of VT100:
// data consist of 3 elements: address, data, byte flag.
func (u *Unibus) accessVT100(byteFlag bool, data ...uint32) error {
	return nil
}
