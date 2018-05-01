package unibus

// Unibus definition
type Unibus struct {

	// Unibus map registers
	UnibusMap [32]int16
}

func (u *Unibus) accessIOPage(physicalAddres uint32, data uint16, byteFlag bool) error {
	return nil
}
