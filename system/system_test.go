package system

import (
	"go/build"
	"os"
	"path/filepath"
	"pdp/console"
	"pdp/unibus"
	"testing"
)

// global resources
var (
	sys *System
	mm  *unibus.MMU18Bit
	c   console.Console
)

// TestMain : initialize memory and CPU
func TestMain(m *testing.M) {
	sys = new(System)
	c = console.NewSimple()
	sys.unibus = unibus.New(&sys.psw, nil, &c, false)
	mm = sys.unibus.Mmu

	sys.unibus.PdpCPU.Reset()
	if err := sys.unibus.Rk01.Attach(0, filepath.Join(build.Default.GOPATH, "src/pdp/rk0")); err != nil {
		panic("Can't mount the drive")
	}
	sys.unibus.Rk01.Reset()

	sys.CPU = sys.unibus.PdpCPU
	sys.CPU.State = unibus.CPURUN

	os.Exit(m.Run())
}

var virtualAddressTests = []struct {
	op             uint16
	virtualAddress uint16
	errorNil       bool
}{
	{0, 0177700, true}, // <- Unibus register address
	{010, 2, true},
	{020, 2, true},
	{030, 1, true},
	{040, 0, true},
	{050, 4, true},
	{061, 020, true},
	{071, 040, true},
}

func TestGetVirtualAddress(t *testing.T) {
	for _, test := range virtualAddressTests {
		// load some value into memory address
		mm.Memory[8] = 040
		mm.Memory[4] = 8
		mm.Memory[2] = 2
		mm.Memory[1] = 1
		mm.Memory[0] = 4
		sys.CPU.Registers[0] = 2

		// setup memory and registers for index mode:
		sys.CPU.Registers[7] = 010
		sys.CPU.Registers[1] = 010

		virtualAddress := sys.CPU.GetVirtualByMode(test.op, 0)
		if virtualAddress != test.virtualAddress {
			t.Errorf("T: %o : Expected virtual address %o got %o\n", test.op, test.virtualAddress, virtualAddress)
		}
	}
}

// try running few lines of machine code
// The memory array is using words, addressing is happening in bytes,
// hence the value pointing to word 0xff is 0x1FE (or 0776 in octal)
func TestRunCode(t *testing.T) {
	sys.CPU.State = unibus.CPURUN

	mm.Memory[0xff] = 2

	code := []uint16{
		012701, // 001000 mov 0xff R1
		000776, // 001002 000377
		062711, // 001004 add 2  to memory pointed by R1 -> mem[0xff]  = 4
		000002, // 001006
		000000, // 001010 done, halt
		000776, // 001012 0377 -> memory address to be loaded to R1
		000002, // 001014 2 -> value to be added
	}

	// load sample code to memory
	memPointer := 001000
	for _, c := range code {
		mm.Memory[memPointer] = c
		memPointer++
	}

	// set PC to starting point:
	sys.CPU.Registers[7] = 002000

	for sys.CPU.State == unibus.CPURUN {
		sys.CPU.Execute()
	}

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
	code := []uint16{
		012700, // 001000 mov 0xff R0
		000377, // 001002 000377 <- value pointed at by R7
		012701, // 001004 mov 0xff R1
		000377, // 001006 0xff <- value pointed at by R7, to be loaded to R1
		// the loop starts here:
		// move the value from R1 to the address pointed by R0
		010120, // 001010 mov R1, (R0)+
		005301, // 001012 dec `R1
		001375, // 001014 BNE -2	<- branch to mov
		000000, // 001016 done, halt
	}

	memPointer := 001000
	for _, c := range code {
		// this should be bytes in 1 word!
		mm.Memory[memPointer] = c & 0xff
		mm.Memory[memPointer+1] = c >> 8
		memPointer += 2
	}

	// set PC to starting point
	sys.CPU.Registers[7] = 001000

	for sys.CPU.State == unibus.CPURUN {
		sys.CPU.Execute()
	}
}

func TestTriggerTrap(t *testing.T) {
	sys.CPU.State = unibus.CPURUN

	code := []uint16{
		066666,
		000000, // 001016 done, halt
	}

	memPointer := 001000
	for _, c := range code {
		// this should be bytes in 1 word!
		mm.Memory[memPointer] = c & 0xff
		mm.Memory[memPointer+1] = c >> 8
		memPointer += 2
	}

	// set PC to starting point
	sys.CPU.Registers[7] = 001000

	for sys.CPU.State == unibus.CPURUN {
		sys.CPU.Execute()
	}
}
