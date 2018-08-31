package pdpcpu

import (
	"bytes"
	"errors"
	"fmt"
	"pdp/interrupts"
	"pdp/mmu"
	"pdp/psw"

	"github.com/jroimartin/gocui"
)

// memory related constans (by far not all needed -- figuring out as while writing)
const (
	// ByteMode -> Read addresses by byte, not by word (?)
	ByteMode = 1

	// ReadMode -> Read from main memory (as opposed to what exactly? (MK))
	ReadMode = 2

	// WriteMode -> Write to main memory
	WriteMode = 4

	// ModifyWord ->  Read and write word in memory
	ModifyWord = ReadMode | WriteMode

	// CPU state: Run / Halt / Wait:
	HALT = 0
	RUN  = 1
	WAIT = 2

	// stack size:
	StackOverflow = 0xff
)

// CPU type:
type CPU struct {
	Registers                   [8]uint16
	floatingPointStatusRegister byte
	psw                         uint16
	State                       int

	// memory access is required:
	// this should be actually managed by unibus, and not here.
	mmunit *mmu.MMU18Bit

	// and stack pointer: kernel, super, illegal, user
	// TODO: Really? -> what is it good for?
	StackPointer [4]uint16

	// track double traps. initialize with false.
	doubleTrap bool

	// original PSW while dealing with trap
	trapPsw psw.PSW

	// trap mask
	trapMask uint16

	// PIR (Programmable Interrupt Register)
	PIR uint16

	// InterruptQueue queue to keep incoming interrupts before processing them
	InterruptQueue []interrupts.Interrupt

	// ClockCounter
	ClockCounter uint16

	// instructions is a map, where key is the opcode,
	// and value is the function executing it
	// the opcode function should append to the following signature:
	// param: instruction int16
	// return: error -> nil if everything went OK
	singleOpOpcodes       map[uint16](func(uint16) error)
	doubleOpOpcodes       map[uint16](func(uint16) error)
	rddOpOpcodes          map[uint16](func(uint16) error)
	controlOpcodes        map[uint16](func(uint16) error)
	singleRegisterOpcodes map[uint16](func(uint16) error)
	otherOpcodes          map[uint16](func(uint16) error)
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
func New(mmunit *mmu.MMU18Bit) *CPU {
	c := CPU{}
	c.mmunit = mmunit
	c.doubleTrap = false
	c.ClockCounter = 0

	// single operand
	c.singleOpOpcodes = make(map[uint16](func(uint16) error))
	c.doubleOpOpcodes = make(map[uint16](func(uint16) error))
	c.rddOpOpcodes = make(map[uint16](func(uint16) error))
	c.controlOpcodes = make(map[uint16](func(uint16) error))
	c.otherOpcodes = make(map[uint16](func(uint16) error))
	c.singleRegisterOpcodes = make(map[uint16](func(uint16) error))

	// single opearnd:
	c.singleOpOpcodes[0100] = c.jmpOp
	c.singleOpOpcodes[0300] = c.swabOp
	c.singleOpOpcodes[05000] = c.clrOp
	c.singleOpOpcodes[0105000] = c.clrbOp
	c.singleOpOpcodes[05100] = c.comOp
	c.singleOpOpcodes[0105100] = c.combOp
	c.singleOpOpcodes[05200] = c.incOp
	c.singleOpOpcodes[0105200] = c.incbOp
	c.singleOpOpcodes[05300] = c.decOp
	c.singleOpOpcodes[0105300] = c.decbOp
	c.singleOpOpcodes[05400] = c.negOp
	c.singleOpOpcodes[0105400] = c.negbOp
	c.singleOpOpcodes[05500] = c.adcOp
	c.singleOpOpcodes[0105500] = c.adcbOp
	c.singleOpOpcodes[05600] = c.sbcOp
	c.singleOpOpcodes[0105600] = c.sbcbOp
	c.singleOpOpcodes[05700] = c.tstOp
	c.singleOpOpcodes[0105700] = c.tstbOp
	c.singleOpOpcodes[06000] = c.rorOp
	c.singleOpOpcodes[0106000] = c.rorbOp
	c.singleOpOpcodes[06100] = c.rolOp
	c.singleOpOpcodes[0106100] = c.rolbOp
	c.singleOpOpcodes[06200] = c.asrOp
	c.singleOpOpcodes[0106200] = c.asrbOp
	c.singleOpOpcodes[06300] = c.aslOp
	c.singleOpOpcodes[0106300] = c.aslbOp
	c.singleOpOpcodes[06400] = c.markOp
	c.singleOpOpcodes[06500] = c.mfpiOp
	c.singleOpOpcodes[06600] = c.mtpiOp
	c.singleOpOpcodes[06700] = c.sxtOp

	// dual operand:
	c.doubleOpOpcodes[010000] = c.movOp
	c.doubleOpOpcodes[0110000] = c.movbOp
	c.doubleOpOpcodes[020000] = c.cmpOp
	c.doubleOpOpcodes[0120000] = c.cmpbOp
	c.doubleOpOpcodes[030000] = c.bitOp
	c.doubleOpOpcodes[0130000] = c.bitbOp
	c.doubleOpOpcodes[040000] = c.bicOp
	c.doubleOpOpcodes[0140000] = c.bicbOp
	c.doubleOpOpcodes[050000] = c.bisOp
	c.doubleOpOpcodes[0150000] = c.bisbOp
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
	c.controlOpcodes[0100000] = c.bplOp // and what the heck happens here??
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
	instruction, err := c.mmunit.ReadMemoryWord(c.Registers[7])
	if err != nil {
		c.trap(interrupts.INTBus)
	}
	c.Registers[7] = (c.Registers[7] + 2) & 0xffff
	return instruction
}

// Decode fetched instruction
// if instruction matching the mask not found in the opcodes map, fallback and try
// to match anything lower.
// Fail ultimately.
func (c *CPU) Decode(instr uint16) func(uint16) error {
	// 2 operand instructions:
	var opcode uint16

	if opcode = instr & 0170000; opcode > 0 {
		if val, ok := c.doubleOpOpcodes[opcode]; ok {
			return val
		}
	}

	// 2 operand instructixon in RDD format
	if opcode = instr & 0177000; opcode > 0 {
		if val, ok := c.rddOpOpcodes[opcode]; ok {
			return val
		}
	}

	// control instructions:
	if opcode = instr & 0177400; opcode > 0 {
		if val, ok := c.controlOpcodes[opcode]; ok {
			return val
		}
	}

	// single operand opcodes
	if opcode = instr & 0177700; opcode > 0 {
		if val, ok := c.singleOpOpcodes[opcode]; ok {
			return val
		}
	}

	// single register opcodes
	if opcode = instr & 0177770; opcode > 0 {
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
func (c *CPU) Execute() {
	if c.State == WAIT {
		return
	}
	instruction := c.Fetch()
	opcode := c.Decode(instruction)
	opcode(instruction)
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
	virtual, err := c.GetVirtualByMode(op, mode)
	if err != nil {
		// TODO: Trigger a trap. something went awry!
		return 0xffff
	}
	data, _ := c.mmunit.ReadMemoryWord(uint16(virtual & 0xffff))
	return data
}

// writeWord writes word value into specified memory address
func (c *CPU) writeWord(op, value uint16) error {
	mode := op >> 3
	register := op & 07

	if mode == 0 {
		c.Registers[register] = value
		return nil
	}
	virtualAddr, err := c.GetVirtualByMode(op, mode)
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

//SetFlag sets CPU carry flag in Processor Status Word
func (c *CPU) SetFlag(flag string, set bool) {
	switch flag {
	case "C":
		c.mmunit.Psw.SetC(set)
	case "V":
		c.mmunit.Psw.SetV(set)
	case "Z":
		c.mmunit.Psw.SetZ(set)
	case "N":
		c.mmunit.Psw.SetN(set)
	case "T":
		c.mmunit.Psw.SetT(set)
	}
}

//GetFlag returns carry flag
func (c *CPU) GetFlag(flag string) bool {
	switch flag {
	case "C":
		return c.mmunit.Psw.C()
	case "V":
		return c.mmunit.Psw.V()
	case "Z":
		return c.mmunit.Psw.Z()
	case "N":
		return c.mmunit.Psw.N()
	case "T":
		return c.mmunit.Psw.T()
	}
	return false
}

// trap handles all trap / abort events.
// TODO: Do I need to signal trap occurence?
func (c *CPU) trap(vector uint16) error {
	if !c.doubleTrap {
		c.trapMask = 0
		c.trapPsw = *c.mmunit.Psw
	} else {
		if c.mmunit.Psw.GetMode() == 0 { // kernel mode
			vector = 4
			c.doubleTrap = true
		}
	}

	// read from kernel D sapce
	// this is valid for pdp 11/44 and 11/70 only. commenting out for now
	//c.mmunit.MMUMode = 0

	newPC, _ := c.mmunit.ReadMemoryWord(vector)
	data, _ := c.mmunit.ReadMemoryWord(vector + 2)
	newPSW := psw.PSW(data)

	// set PREVIOUS MODE bits in new PSW -> take it from currentMode bits in
	// saved c.trapPSW
	newPSW = (newPSW & 0xcfff) | ((c.trapPsw >> 2) & 0x3000)

	// set new Processor Status Word
	c.mmunit.Psw = &newPSW

	// TODO: - Double Trap not implemented

	// set new Program counter:
	c.Registers[7] = newPC

	c.doubleTrap = false
	return nil
}

// PopWord pops 1 word from Processor stack:
func (c *CPU) PopWord() uint16 {
	result, _ := c.mmunit.ReadMemoryWord(c.Registers[6])

	// update Stack Pointer after reading the word
	c.Registers[6] = (c.Registers[6] + 2) & 0xffff

	return result
}

// GetVirtualByMode returns virtual address extracted from the CPU instuction
func (c *CPU) GetVirtualByMode(instruction, accessMode uint16) (uint16, error) {
	var addressInc uint16
	reg := instruction & 7
	addressMode := (instruction >> 3) & 7
	var virtAddress uint16

	switch addressMode {
	case 0:
		// TODO: REALLY throw a trap here.
		return 0, errors.New("Wrong address mode - throw trap?")
	case 1:
		// register keeps the address:
		virtAddress = c.Registers[reg]
	case 2:
		// register keeps the address. Increment the value by 2 (word!)
		// TODO: value should be incremented by 1 if byte instruction used.
		addressInc = 2
		virtAddress = c.Registers[reg]
		c.Registers[reg] = (c.Registers[reg] + addressInc) & 0xffff
	case 3:
		// autoincrement deferred
		// TODO: ADD special cases (R6 and R7)
		addressInc = 2
		virtAddress = c.Registers[reg]
		c.Registers[reg] = (c.Registers[reg] + addressInc) & 0xffff
	case 4:
		// autodecrement - step depends on which register is in use:
		addressInc = 2
		if (reg < 6) && (accessMode&ByteMode > 0) {
			addressInc = 1
		}
		virtAddress = (c.Registers[reg] + addressInc) & 0xffff
		c.Registers[reg] = (c.Registers[reg] - addressInc) & 0xffff
	case 5:
		// autodecrement deferred
		virtAddress = (c.Registers[reg] - 2) & 0xffff
	case 6:
		// index mode -> read next word to get the basis for address, add value in Register
		baseAddr, _ := c.mmunit.ReadMemoryWord(c.Registers[7])
		virtAddress = (baseAddr + c.Registers[reg]) & 0xffff

		// increment program counter register
		c.Registers[7] = (c.Registers[7] + 2) & 0xffff
	case 7:
		baseAddr, _ := c.mmunit.ReadMemoryWord(c.Registers[7])
		virtAddress = (baseAddr + c.Registers[reg]) & 0xffff
		virtAddress, _ = c.mmunit.ReadMemoryWord(virtAddress)
		// increment program counter register
		c.Registers[7] = (c.Registers[7] + 2) & 0xffff
	}
	// all-catcher return
	return virtAddress, nil
}
