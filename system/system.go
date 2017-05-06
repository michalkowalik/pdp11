package system

import (
	"fmt"
	"pdp/pdpcpu"
)

// System definition.
type System struct {
	Memory [4 * 1024 * 1024]byte
	CPU    *pdpcpu.CPU

	// Unibus map registers
	UnibusMap [32]int16
}

// InitializeSystem initializes the emulated PDP-11/44 hardware
func InitializeSystem() *System {
	sys := new(System)
	sys.CPU = new(pdpcpu.CPU)
	fmt.Printf("Initializing PDP11 CPU...\n")
	return sys
}

// Noop is a dummy function just to keep go compiler happy for a while
func (sys *System) Noop() {
	fmt.Printf(".. Noop ..\n")
}
