package system

import (
	"pdp/mmu"
	"pdp/pdpcpu"
	"testing"
)

func TestRegisterRead(t *testing.T) {
	sys := new(System)
	mmunit = mmu.MMU{}
	mmunit.Memory = &sys.Memory

	sys.CPU = pdpcpu.New(&mmunit)

	// load an address to register
	sys.CPU.Registers[0] = 2
	expected := " |R0: 02 |  |R1: 0 |  |R2: 0 |  |R3: 0 |  |R4: 0 |  |R5: 0 |  |R6: 0 |  |R7: 0 | "
	returned := sys.CPU.PrintRegisters()
	if returned != expected {
		t.Error("Expected value: >", expected, "<, got: >", returned, "<")
	}
}
