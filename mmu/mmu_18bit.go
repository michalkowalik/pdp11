package mmu

// MMU18Bit implements the 18 bit memory management unit
// max 256KB RAM (128K words), enough to run early unix versions
type MMU18Bit struct {

	// Memory : Physical memory
	Memory *[128 * 1024]uint16

	// PAR : Page Address Registers
	PAR [16]uint16

	// PDR : Page Description Registers
	PDR [16]uint16

	// APR : Active Page Register - 8 of them
	APR [8]uint16
}
