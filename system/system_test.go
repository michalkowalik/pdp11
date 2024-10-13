package system

import (
	"go/build"
	"os"
	"path/filepath"
	"pdp/console"
	"pdp/interrupts"
	"pdp/psw"
	"pdp/unibus"
	"testing"
)

// global resources
var (
	sys *System
	mm  unibus.MMU
	c   console.Console
)

// TestMain : initialize memory and CPU
func TestMain(m *testing.M) {
	sys = new(System)
	c = console.NewSimple()
	sys.unibus = unibus.New(&sys.psw, nil, &c, false)
	mm = sys.unibus.Mmu

	sys.unibus.PdpCPU.Reset()
	if err := sys.unibus.Rk01.Attach(0, filepath.Join(build.Default.GOPATH, "src/pdp11/rk0")); err != nil {
		//if err := sys.unibus.Rk01.Attach(0, filepath.Join("/Users/mkowalik", "src/pdp11/rk0")); err != nil {
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
		sys.unibus.Memory[8] = 040
		sys.unibus.Memory[4] = 8
		sys.unibus.Memory[2] = 2
		sys.unibus.Memory[1] = 1
		sys.unibus.Memory[0] = 4
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

	sys.unibus.Memory[0xff] = 2

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
		sys.unibus.Memory[memPointer] = c
		memPointer++
	}

	// set PC to starting point:
	sys.CPU.Registers[7] = 002000

	for sys.CPU.State == unibus.CPURUN {
		sys.CPU.Execute()
	}

	if memVal := sys.unibus.Memory[0xff]; memVal != 4 {
		t.Errorf("Expected memory cell at 0xff to be equal 4, got %x\n", memVal)
	}
}

// another try of running a bit of assembler code
// this time with branch instructions
// the instruction is to start at memory address 0xff
// and fill the next 256 memory addresses with increasing values
// bne should break the loop
// TODO: Assertions for the code
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
		sys.unibus.Memory[memPointer] = c & 0xff
		sys.unibus.Memory[memPointer+1] = c >> 8
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
		sys.unibus.Memory[memPointer] = c & 0xff
		sys.unibus.Memory[memPointer+1] = c >> 8
		memPointer += 2
	}

	// set PC to starting point
	sys.CPU.Registers[7] = 001000

	for sys.CPU.State == unibus.CPURUN {
		sys.CPU.Execute()
	}
}

func TestInterruptHandling(t *testing.T) {
	sys.CPU.KernelStackPointer = 0777 << 1
	sys.CPU.UserStackPointer = 01776 << 1

	memPointer := uint16(04000)
	r7Value := uint16(0xfffe)
	initialPSW := uint16(0xf000)

	sys.unibus.Memory[memPointer>>1] = 06 // RTI

	sys.unibus.Memory[interrupts.INTRK>>1] = memPointer

	sys.CPU.Registers[7] = r7Value
	sys.CPU.Registers[6] = sys.CPU.KernelStackPointer

	sys.CPU.SwitchMode(psw.UserMode)
	// mode = user, previousMode = user, no flags.
	sys.unibus.WriteIO(unibus.PSWAddr, initialPSW)

	sys.unibus.SendInterrupt(4, interrupts.INTRK)
	if sys.unibus.InterruptQueue[0].Vector != interrupts.INTRK {
		t.Errorf("Expected to have INTRK in the interrupt queue")
	}

	sys.processInterrupt(sys.unibus.InterruptQueue[0])

	if sys.unibus.Psw.GetMode() != unibus.KernelMode {
		t.Errorf("Expected processor to be in kernel mode")
	}

	if (sys.unibus.Psw.Get()>>12)&3 != unibus.UserMode {
		t.Errorf("Expected previousMode to be USER")
	}

	if sys.CPU.Registers[6]>>1 != 0775 { // r7 and original psw should be on the stack
		t.Errorf("Expected kernel stack pointer to be pointing to the address of 0775, got %o\n", sys.CPU.Registers[6])
	}

	instruction := sys.CPU.Fetch()
	if instruction != 06 {
		t.Errorf("Expected to fetch RTI at this point, but got %o\n", instruction)
	}

	// execute RTI
	(sys.CPU.Decode(instruction))(instruction)

	if sys.CPU.Registers[7] != r7Value {
		t.Errorf("Expected SP to be set back to the original value, but got %o\n", sys.CPU.Registers[7])
	}

	// previous mode should be set to kernel
	if sys.unibus.Psw.Get() != initialPSW {
		t.Errorf("Expected PSW to be set to the original value, but got %x\n", sys.unibus.Psw.Get())
	}

	if sys.CPU.Registers[6] != sys.CPU.UserStackPointer {
		t.Errorf("Stack pointer should be set to the user stack by now")
	}
}
