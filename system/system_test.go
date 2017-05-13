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

var virtualAddressTests = []struct {
	op             uint16
	virtualAddress uint32
	errorNil       bool
}{
	{0, 0, false},
	{010, 2, true},
	{020, 2, true},
	{030, 2, true},
	{040, 0, true}, // <- autodecrement! expect dragons! and re-test with byte mode
	{050, 0, true},
}

// check if an address in memory can be read
func TestGetVirtualAddress(t *testing.T) {
	sys := new(System)
	mmunit = mmu.MMU{}
	mmunit.Memory = &sys.Memory
	sys.CPU = pdpcpu.New(&mmunit)

	for _, test := range virtualAddressTests {
		// load some value into memory address
		sys.Memory[2] = 2
		sys.Memory[1] = 1
		sys.Memory[0] = 4
		sys.CPU.Registers[0] = 2

		virtualAddress, err := mmunit.GetVirtualByMode(&sys.CPU.Registers, test.op, 0)
		if virtualAddress != test.virtualAddress {
			t.Logf("Registers: %s\n", sys.CPU.PrintRegisters())
			t.Error("Expected virtual address ", test.virtualAddress, " , got ", virtualAddress)
		}
		if (err == nil) != test.errorNil {
			t.Errorf("Unexpected error value: %v. Expected: nil\n", err)
		}
	}
}
