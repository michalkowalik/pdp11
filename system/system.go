package system

import (
	"fmt"
	"pdp/pdpcpu"
)

// system definition.
// Memory - static array, 4MB in size:
var memory [4 * 1024 * 1024]byte
var cpu *pdpcpu.CPU

// InitializeSystem initializes the emulated PDP-11/44 hardware
func InitializeSystem() {
	cpu = new(pdpcpu.CPU)
	fmt.Printf("Initializing PDP11 ...\n")
}
