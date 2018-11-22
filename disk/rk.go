package disk

// RK11 disk controller
type RK11 struct {
	// Registers -> check description in attached markdown
	RKDS uint16
	RKER uint16
	RKCS uint16
	RKWC uint16
	RKBA uint16

	// disk units
	unit [8]*RK05

	//disk geometry:
	// TODO: make sure "int" is the most fitting type. Perhaps uint16 is good enough.
	drive, sector, surface, cylinder int

	running bool

	// we also somehow need the unibus here...
	// think about channel communication with unibus.
}

// RK05 disk cartridge
type RK05 struct {
	rdisk  []byte
	locked bool
}

// Attach reads disk image file and loads it to memory
func (r *RK11) Attach(drive int, path string) error {
	return nil
}
