package pdpcpu

import (
	"bytes"
	"fmt"
	"pdp/mmu"

	"github.com/jroimartin/gocui"
)

// memory related constans (by far not all needed -- figuring out as while writing)
const (
	// ByteMode -> Read addresses by byte, not by word (?)
	ByteMode = 1

	// ReadMode -> Read from main memory
	ReadMode = 2

	// WriteMode -> Write from main memory
	WriteMode = 4

	// ModifyWord ->  Read and write word in memory
	ModifyWord = ReadMode | WriteMode

	// CPU state: Run / Halt:
	HALT = 0
	RUN  = 1
)

// CPU type:
type CPU struct {
	Registers                   [8]uint16
	statusFlags                 byte // not needed?
	floatingPointStatusRegister byte
	statusRegister              uint16
	State                       int

	// memory access is required:
	mmunit *mmu.MMU

	// instructions is a map, where key is the opcode,
	// and value is the function executing it
	// the opcode function should append to the following signature:
	// param: instruction int16
	// return: error -> nil if everything went OK
	singleOpOpcodes       map[uint16](func(int16) error)
	doubleOpOpcodes       map[uint16](func(int16) error)
	rddOpOpcodes          map[uint16](func(int16) error)
	controlOpcodes        map[uint16](func(int16) error)
	singleRegisterOpcodes map[uint16](func(int16) error)
	otherOpcodes          map[uint16](func(int16) error)
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

//New initializes and returns the CPU variable:
func New(mmunit *mmu.MMU) *CPU {
	c := CPU{}
	c.mmunit = mmunit
	// single operand
	c.singleOpOpcodes = make(map[uint16](func(int16) error))
	c.doubleOpOpcodes = make(map[uint16](func(int16) error))
	c.rddOpOpcodes = make(map[uint16](func(int16) error))
	c.controlOpcodes = make(map[uint16](func(int16) error))
	c.otherOpcodes = make(map[uint16](func(int16) error))
	c.singleRegisterOpcodes = make(map[uint16](func(int16) error))

	// single opearnd:
	c.singleOpOpcodes[05000] = c.clrOp // check if it OK?
	c.singleOpOpcodes[0100] = c.jmpOp
	c.singleOpOpcodes[06400] = c.markOp
	c.singleOpOpcodes[06500] = c.mfpiOp
	c.singleOpOpcodes[06600] = c.mtpiOp

	// dual operand:
	c.doubleOpOpcodes[010000] = c.movOp
	c.doubleOpOpcodes[020000] = c.cmpOp
	c.doubleOpOpcodes[030000] = c.bitOp
	c.doubleOpOpcodes[040000] = c.bicOp
	c.doubleOpOpcodes[050000] = c.bisOp
	c.doubleOpOpcodes[060000] = c.addOp
	c.doubleOpOpcodes[0160000] = c.subOp

	// RDD dual operand:
	c.rddOpOpcodes[070000] = c.mulOp
	c.rddOpOpcodes[071000] = c.divOp
	c.rddOpOpcodes[072000] = c.ashOp
	c.rddOpOpcodes[073000] = c.ashcOp
	c.rddOpOpcodes[074000] = c.xorOp
	c.rddOpOpcodes[04000] = c.jsrOp
	c.rddOpOpcodes[077000] = c.sobOp

	// control instructions & traps:
	c.controlOpcodes[0400] = c.brOp
	c.controlOpcodes[01000] = c.bneOp
	c.controlOpcodes[01400] = c.beqOp
	c.controlOpcodes[0100000] = c.bplOp
	c.controlOpcodes[0100400] = c.bmiOp
	c.controlOpcodes[0102000] = c.bvsOp
	c.controlOpcodes[0103000] = c.bccOp
	c.controlOpcodes[0103400] = c.bcsOp
	c.controlOpcodes[0104000] = c.emtOp
	c.controlOpcodes[0104400] = c.trapOp

	// conditional branching - signed int
	c.controlOpcodes[02000] = c.bgeOp
	c.controlOpcodes[02400] = c.bltOp
	c.controlOpcodes[03000] = c.bgtOp
	c.controlOpcodes[03400] = c.bleOp

	// conditional branching - unsigned int
	c.controlOpcodes[0101000] = c.bhiOp
	c.controlOpcodes[0101400] = c.blosOp
	c.controlOpcodes[0103000] = c.bhisOp
	c.controlOpcodes[0103400] = c.bloOp

	// single register & condition code opcodes
	c.singleRegisterOpcodes[0200] = c.rtsOp
	c.singleRegisterOpcodes[0240] = c.setFlagOp
	c.singleRegisterOpcodes[0250] = c.setFlagOp
	c.singleRegisterOpcodes[0260] = c.clearFlagOp
	c.singleRegisterOpcodes[0270] = c.clearFlagOp

	// no operand:
	c.otherOpcodes[0] = c.haltOp
	c.otherOpcodes[1] = c.waitOp
	c.otherOpcodes[2] = c.rtiOp
	c.otherOpcodes[3] = c.bptOp
	c.otherOpcodes[4] = c.iotOp
	c.otherOpcodes[5] = c.resetOp
	c.otherOpcodes[6] = c.rttOp
	return &c
}

// cpu should be able to fetch, decode and execute:

// Fetch next instruction from memory
// Address to fetch is kept in R7 (PC)
func (c *CPU) Fetch() uint16 {
	instruction := c.mmunit.ReadMemoryWord(c.Registers[7])
	c.Registers[7] = (c.Registers[7] + 2) & 0xffff
	return instruction
}

// Decode fetched instruction
// if instruction matching the mask not found in the opcodes map, fallback and try
// to match anything lower.
// Fail ultimately.
func (c *CPU) Decode(instr uint16) func(int16) error {
	// 2 operand instructions:
	if opcode := instr & 0170000; opcode > 0 {
		if val, ok := c.doubleOpOpcodes[opcode]; ok {
			return val
		}
	}

	// 2 operand instructixon in RDD format
	if opcode := instr & 0177000; opcode > 0 {
		if val, ok := c.rddOpOpcodes[opcode]; ok {
			return val
		}
	}

	// control instructions:
	if opcode := instr & 0177400; opcode > 0 {
		if val, ok := c.controlOpcodes[opcode]; ok {
			return val
		}
	}

	// single operand opcodes
	if opcode := instr & 0177700; opcode > 0 {
		if val, ok := c.singleOpOpcodes[opcode]; ok {
			return val
		}
	}

	// single register opcodes
	if opcode := instr & 0177770; opcode > 0 {
		if val, ok := c.singleRegisterOpcodes[opcode]; ok {
			return val
		}
	}

	// TODO: add "if debug"
	// fmt.Printf("opcode: %#o\n", opcode)

	// everything else:
	return c.otherOpcodes[instr]

}

// Execute decoded instruction
func (c *CPU) Execute() error {
	instruction := c.Fetch()
	opcode := c.Decode(instruction)
	return opcode(int16(instruction))
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
