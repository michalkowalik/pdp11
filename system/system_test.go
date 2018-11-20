package system

import (
	"pdp/psw"
	"pdp/unibus"
	"testing"
)

// global resources
var (
	sys *System
	mm  unibus.MMU18Bit
)

// TestMain : initialize memory and CPU
func TestMain(m *testing.M) {
	sys = new(System)
	mm = unibus.MMU18Bit{}
	p := psw.PSW(0)
	mm.Psw = &p
	sys.CPU = unibus.NewCPU(&mm)
}

var virtualAddressTests = []struct {
	op             uint16
	virtualAddress uint16
	errorNil       bool
}{
	{0, 0, false},
	{010, 2, true},
	{020, 2, true},
	{030, 2, true},
	{040, 4, true}, // <- autodecrement! expect dragons! and re-test with byte mode
	{050, 0, true},
	{061, 020, true}, // <- err!! err!!
	{071, 040, true},
}

// check if an address in memory can be read
func TestGetVirtualAddress(t *testing.T) {
	for _, test := range virtualAddressTests {
		// load some value into memory address
		mm.Memory[2] = 2
		mm.Memory[1] = 1
		mm.Memory[0] = 4
		sys.CPU.Registers[0] = 2

		// setup memory and registers for index mode:
		sys.CPU.Registers[7] = 010
		sys.CPU.Registers[1] = 010
		mm.Memory[010] = 010
		mm.Memory[020] = 040

		virtualAddress, err := sys.CPU.GetVirtualByMode(test.op, 0)
		if virtualAddress != test.virtualAddress {
			t.Error("Expected virtual address ", test.virtualAddress, " , got ", virtualAddress)
		}
		if (err == nil) != test.errorNil {
			t.Errorf("Unexpected error value: %v. Expected: nil\n", err)
		}
	}
}

// try running few lines of machine code
func TestRunCode(t *testing.T) {
	sys.CPU.State = unibus.CPURUN

	mm.Memory[0xff] = 2

	code := []uint16{
		012701, // 001000 mov 0xff R1
		000377, // 001002 000377
		062711, // 001004 add 2  to memory pointed by R1 -> mem[0xff]  = 4
		000002, // 001006
		000000, // 001010 done, halt
		000377, // 001012 0377 -> memory address to be loaded to R1
		000002, // 001014 2 -> value to be added
	}

	// load sample code to memory
	memPointer := 001000
	for _, c := range code {

		// this should be actually bytes in word!
		mm.Memory[memPointer] = uint16(c & 0xff)
		mm.Memory[memPointer+1] = uint16(c >> 8)
		memPointer += 2
	}

	// set PC to starting point:
	sys.CPU.Registers[7] = 001000

	for sys.CPU.State == unibus.CPURUN {
		sys.CPU.Execute()
	}

	// assert memory at 0xff = 4
	if memVal := mm.Memory[0xff]; memVal != 4 {
		t.Errorf("Expected memory cell at 0xff to be equal 4, got %x\n", memVal)
	}
}

// another try of running a bit of assembler code
// this time with branch instructions
// the instruction is to start at memory address 0xff
// and fill the next 256 memory addresses with increasing values
// bne should break the loop
func TestRunBranchCode(t *testing.T) {
	sys.CPU.State = unibus.CPURUN
	// sample code
	code := []uint16{
		012700, // 001000 mov 0xff R0
		000377, // 001002 000377 <- value pointed at by R7
		012701, // 001004 mov 0xff R1
		000377, // 001006 0xff <- value pointed at by R7, to be loaded to R1
		//// the loop starts here:
		//// move the value from R1 to the address pointed by R0
		010120, // 001010 mov R1, (R0)+
		005301, // 001012 dec `R1
		001375, // 001014 BNE -2	<- branch to mov
		000000, // 001016 done, halt
	}

	// load sample code to memory
	memPointer := 001000
	for _, c := range code {

		// this should be bytes in 1 word!
		mm.Memory[memPointer] = uint16(c & 0xff)
		mm.Memory[memPointer+1] = uint16(c >> 8)
		memPointer += 2
	}

	// set PC to starting point
	sys.CPU.Registers[7] = 001000

	for sys.CPU.State == unibus.CPURUN {
		sys.CPU.Execute()
	}

	// assert results
}

func TestTriggerTrap(t *testing.T) {
	sys.CPU.State = unibus.CPURUN

	code := []uint16{
		066666,
		000000, // 001016 done, halt
	}

	// load sample code to memory
	memPointer := 001000
	for _, c := range code {

		// this should be bytes in 1 word!
		mm.Memory[memPointer] = uint16(c & 0xff)
		mm.Memory[memPointer+1] = uint16(c >> 8)
		memPointer += 2
	}

	// set PC to starting point
	sys.CPU.Registers[7] = 001000

	for sys.CPU.State == unibus.CPURUN {
		sys.CPU.Execute()
	}
}
