package pdpcpu

import (
	"bytes"
	"fmt"
	"pdp/mmu"

	"github.com/jroimartin/gocui"
)

// CPU type:
type CPU struct {
	Registers                   [8]uint16
	statusFlags                 byte // not needed?
	floatingPointStatusRegister byte
	statusRegister              uint16

	// memory access is required:
	mmunit *mmu.MMU

	// instructions is a map, where key is the opcode,
	// and value is the function executing it
	// the opcode function should append to the following signature:
	// param: instruction int16
	// return: error -> nil if everything went OK
	opcodes map[uint16](func(int16) error)
}

/**
* processor flags (or Condition Codes, as PDP-11 Processor Handbook wants it)
* C -> Carry flag
* V -> Overflow
* Z -> Zero
* N -> Negative
* T -> Trap
 */
var cpuFlags = map[string]struct {
	setMask   uint16
	unsetMask uint16
}{
	"C": {1, 0xfffe},
	"V": {2, 0xfffd},
	"Z": {4, 0xfffb},
	"N": {8, 0xfff7},
	"T": {0x10, 0xffef},
}

// memory related constans (by far not all needed -- figuring out as while writing)

// ByteMode -> Read addresses by byte, not by word (?)
const ByteMode = 1

// ReadMode -> Read from main memory
const ReadMode = 2

// WriteMode -> Write from main memory
const WriteMode = 4

// ModifyWord ->  Read and write word in memory
const ModifyWord = ReadMode | WriteMode

//New initializes and returns the CPU variable:
func New(mmunit *mmu.MMU) *CPU {
	c := CPU{}
	c.mmunit = mmunit
	// single operand
	c.opcodes = make(map[uint16](func(int16) error))

	// single opearnd:
	c.opcodes[050] = c.clrOp

	// dual operand:
	c.opcodes[01] = c.movOp
	c.opcodes[06] = c.addOp
	return &c
}

// cpu should be able to fetch, decode and execute:

// Fetch next instruction from memory
func (c *CPU) Fetch() {
	fmt.Printf("CPU Fetch\n")
}

//Decode fetched instruction
func (c *CPU) Decode(instr uint16) func(int16) error {
	var opcode uint16
	// 2 operand instructions:
	if (instr & 0170000) > 0 {
		opcode = instr >> 9
	}
	// single operand
	opcode = instr >> 6
	return c.opcodes[opcode]

}

// Execute decoded instruction
func (c *CPU) Execute() {
	fmt.Printf("Execute.. \n")
}

// helper functions:

// readWord returns value specified by source or destination part of the operand.
func (c *CPU) readWord(op uint16) uint16 {
	// check mode:
	mode := op >> 3
	register := op & 07
	if mode == 0 {
		//value directly in register
		return c.Registers[register]
	}
	// TODO: access mode is hardcoded to 1 !! <- needs to be changed or removed
	virtual, err := c.mmunit.GetVirtualByMode(&c.Registers, op, 1)
	if err != nil {
		// TODO: Trigger a trap. something went awry!
		return 0xffff
	}
	return c.mmunit.ReadMemoryWord(uint16(virtual & 0xffff))
}

// writeWord writes word value into specified memory address
func (c *CPU) writeWord(op, value uint16) error {
	mode := op >> 3
	register := op & 07

	if mode == 0 {
		c.Registers[register] = value
		return nil
	}
	virtualAddr, err := c.mmunit.GetVirtualByMode(&c.Registers, op, 1)
	if err != nil {
		return err
	}
	c.mmunit.WriteMemoryWord(uint16(virtualAddr&0xffff), value)
	return nil
}

//PrintRegisters returns buffer status as a string
func (c *CPU) PrintRegisters() string {
	var buffer bytes.Buffer
	for i, reg := range c.Registers {
		buffer.WriteString(fmt.Sprintf(" |R%d: %#o | ", i, reg))
	}
	return buffer.String()
}

// DumpRegisters displays register values
func (c *CPU) DumpRegisters(regView *gocui.View) {
	for i, reg := range c.Registers {
		fmt.Fprintf(regView, " |R%d: %#o | ", i, reg)
	}
}

// status word handling:

//SetFlag sets CPU carry flag in Processor Status Word
func (c *CPU) SetFlag(flag string, set bool) {
	if set == true {
		c.statusRegister = c.statusRegister | cpuFlags[flag].setMask
	} else {
		c.statusRegister = c.statusRegister & cpuFlags[flag].unsetMask
	}
}

//GetFlag returns carry flag
func (c *CPU) GetFlag(flag string) bool {
	if c.statusRegister&cpuFlags[flag].setMask != 0 {
		return true
	}
	return false
}
