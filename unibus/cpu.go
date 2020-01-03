package unibus

import (
	"errors"
	"fmt"
	"pdp/interrupts"
	"pdp/psw"
	"strings"
)

// memory related constans (by far not all needed -- figuring out as while writing)
const (
	// add debug output to the console
	debug = true

	// ByteMode -> Read addresses by byte, not by word (?)
	ByteMode = 1

	// ReadMode -> Read from main memory (as opposed to what exactly? (MK))
	ReadMode = 2

	// WriteMode -> Write to main memory
	WriteMode = 4

	// ModifyWord ->  Read and write word in memory
	ModifyWord = ReadMode | WriteMode

	// CPU state: Run / Halt / Wait:
	HALT   = 0
	CPURUN = 1
	WAIT   = 2

	// stack size:
	StackOverflow = 0xff
)

// CPU type:
type CPU struct {
	Registers                   [8]uint16
	floatingPointStatusRegister byte
	State                       int

	// system stack pointers: kernel, super, illegal, user
	// super won't be needed for pdp11/40:
	KernelStackPointer uint16
	UserStackPointer   uint16

	// memory access is required:
	// this should be actually managed by unibus, and not here.
	mmunit *MMU18Bit

	// original PSW while dealing with trap
	trapPsw psw.PSW

	// trap mask
	trapMask uint16

	// PIR (Programmable Interrupt Register)
	PIR uint16

	// ClockCounter
	ClockCounter uint16

	// instructions is a map, where key is the opcode,
	// and value is the function executing it
	// the opcode function should append to the following signature:
	// param: instruction int16
	// return: error -> nil if everything went OK
	singleOpOpcodes       map[uint16](func(uint16))
	doubleOpOpcodes       map[uint16](func(uint16))
	rddOpOpcodes          map[uint16](func(uint16))
	controlOpcodes        map[uint16](func(uint16))
	singleRegisterOpcodes map[uint16](func(uint16))
	otherOpcodes          map[uint16](func(uint16))
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

//NewCPU initializes and returns the CPU variable:
func NewCPU(mmunit *MMU18Bit) *CPU {

	c := CPU{}
	c.mmunit = mmunit
	c.ClockCounter = 0

	// single operand
	c.singleOpOpcodes = make(map[uint16](func(uint16)))
	c.doubleOpOpcodes = make(map[uint16](func(uint16)))
	c.rddOpOpcodes = make(map[uint16](func(uint16)))
	c.controlOpcodes = make(map[uint16](func(uint16)))
	c.otherOpcodes = make(map[uint16](func(uint16)))
	c.singleRegisterOpcodes = make(map[uint16](func(uint16)))

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
	defer func() {
		t := recover()
		switch t := t.(type) {
		case interrupts.Trap:
			c.Trap(t.Vector)
		case nil:
			// ignore
		default:
			panic(t)
		}
	}()

	instruction := c.mmunit.ReadMemoryWord(c.Registers[7])
	c.Registers[7] = (c.Registers[7] + 2) & 0xffff
	return instruction
}

// Decode fetched instruction
// if instruction matching the mask not found in the opcodes map, fallback and try
// to match anything lower.
// Fail ultimately.
func (c *CPU) Decode(instr uint16) func(uint16) {
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

	if opcode = instr & 07; opcode > 0 {
		if val, ok := c.otherOpcodes[opcode]; ok {
			return val
		}
	}

	// haltOp has optcode of 0, easiest to treat it separately
	if instr == 0 {
		return c.otherOpcodes[0]
	}

	// at this point it can be only an invalid instruction:
	fmt.Printf(c.printState(instr))
	fmt.Printf("%s\n", c.mmunit.unibus.Disasm(instr))
	fmt.Printf("\nInstruction : %o\n", instr)
	panic(interrupts.Trap{Vector: interrupts.INTInval, Msg: "Invalid Instruction"})

}

// Execute decoded instruction
func (c *CPU) Execute() {
	instruction := c.Fetch()
	opcode := c.Decode(instruction)

	if debug {
		fmt.Printf(c.printState(instruction))
		fmt.Printf("%s\n", c.mmunit.unibus.Disasm(instruction))

	}
	opcode(instruction)
}

// helper functions:

// readMemory reads either a word or a byte from memory
// mode: 1: byte, 0: word
func (c *CPU) readFromMemory(op uint16, length uint16) uint16 {
	var byteData byte
	var data uint16
	var err error

	defer func() {
		t := recover()
		switch t := t.(type) {
		case interrupts.Trap:
			fmt.Printf("Triggering trap in readFromMemory")
			c.Trap(t.Vector)
		case nil:
			// ignore
		default:
			panic(t)
		}
	}()

	// check mode:
	mode := (op >> 3) & 7
	register := op & 7

	if mode == 0 {

		//value directly in register
		if length == 0 {
			return c.Registers[register]
		}

		// and for the byte mode:
		return c.Registers[register] & 0xFF
	}
	virtual, err := c.GetVirtualByMode(op, length)
	if err != nil {
		panic("Can't obtain virtual address")
	}

	if length == 0 {
		data = c.mmunit.ReadMemoryWord(virtual)
	} else {
		byteData = c.mmunit.ReadMemoryByte(virtual)
		data = uint16(byteData)
	}
	return data
}

// readWord returns value specified by source or destination part of the operand.
func (c *CPU) readWord(op uint16) uint16 {
	return c.readFromMemory(op, 0)
}

// read byte
func (c *CPU) readByte(op uint16) byte {
	return byte(c.readFromMemory(op, 1))
}

// writeMemory writes either byte or word,
// complementary to read operations
func (c *CPU) writeMemory(op, value, length uint16) error {

	defer func() {
		t := recover()
		switch t := t.(type) {
		case interrupts.Trap:
			fmt.Printf("Triggering trap in writeMemory")
			c.Trap(t.Vector)
		case nil:
			// ignore
		default:
			panic(t)
		}
	}()

	mode := (op >> 3) & 7
	register := op & 7

	// TODO: clean it up -> it is awful hack, and its
	// not what DEC says!
	if mode == 0 {
		if length == 1 {
			c.Registers[register] = (value & 0xFF)
		} else {
			c.Registers[register] = value
		}
		return nil
	}
	virtualAddr, err := c.GetVirtualByMode(op, length)
	if err != nil {
		return err
	}
	if length == 0 {
		c.mmunit.WriteMemoryWord(virtualAddr, value)
	} else {
		c.mmunit.WriteMemoryByte(virtualAddr, byte(value))
	}
	return nil
}

// writeWord writes word value into specified memory address
func (c *CPU) writeWord(op, value uint16) error {
	return c.writeMemory(op, value, 0)
}

// writeByte writes byte value into specified memory location
func (c *CPU) writeByte(op, value uint16) error {
	return c.writeMemory(op, value, 1)
}

// DumpRegisters displays register values
func (c *CPU) DumpRegisters() string {
	var res strings.Builder
	for i, reg := range c.Registers {
		fmt.Fprintf(&res, "R%d %06o ", i, reg)
	}
	s := res.String()
	return s[:(len(s) - 1)]
}

func (c *CPU) printState(instruction uint16) string {
	//registers
	out := fmt.Sprintf("%s\n", c.DumpRegisters())

	// flags
	out += fmt.Sprintf("%s ", c.mmunit.unibus.psw.GetFlags())

	// instruction
	out += fmt.Sprintf(" instr %06o: %06o   ", c.Registers[7]-2, instruction)

	return out
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

// SwitchMode switches the kernel / user mode:
// 0 for user, 3 for kernel, everything else is a mistake.
// values are as they are used in the PSW
func (c *CPU) SwitchMode(m uint16) {
	c.mmunit.Psw.SwitchMode(m)

	// save processor stack pointers:
	if m > 0 {
		c.UserStackPointer = c.Registers[6]
	} else {
		c.KernelStackPointer = c.Registers[6]
	}

	// set processor stack:
	if m > 0 {
		c.Registers[6] = c.UserStackPointer
	} else {
		c.Registers[6] = c.KernelStackPointer
	}
}

// Trap handles all Trap / abort events.
// strikingly similar to processInterrupt.
func (c *CPU) Trap(vector uint16) {
	prev := c.mmunit.Psw.Get()
	defer func(prev uint16) {
		t := recover()
		switch t := t.(type) {
		case interrupts.Trap:
			panic("Red stack trap. Fatal.")
		case nil:
			break
		default:
			panic(t)
		}
		c.Registers[7] = c.mmunit.ReadMemoryWord(vector)
		intPSW := c.mmunit.ReadMemoryWord(vector + 2)

		if (prev & (1 << 14)) > 0 {
			intPSW |= (1 << 13) | (1 << 12)
		}
		c.mmunit.Psw.Set(intPSW)
	}(prev)

	if vector&1 == 1 {
		panic("Odd vector number!")
	}

	c.SwitchMode(psw.KernelMode)
	c.Push(prev)
	c.Push(c.Registers[7])
}

// GetVirtualByMode returns virtual address extracted from the CPU instuction
// access mode: 0 for Word, 1 for Byte
func (c *CPU) GetVirtualByMode(instruction, accessMode uint16) (uint16, error) {
	addressInc := uint16(2)
	reg := instruction & 7
	addressMode := (instruction >> 3) & 7
	var virtAddress uint16

	// byte mode
	if accessMode == 1 {
		addressInc = 1
	}

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
		virtAddress = c.Registers[reg]
		c.Registers[reg] = (c.Registers[reg] + addressInc) & 0xffff
	case 3:
		// autoincrement deferred --> it doesn't look like byte mode applies here?
		virtAddress = c.mmunit.ReadMemoryWord(c.Registers[reg])
		c.Registers[reg] = (c.Registers[reg] + 2) & 0xffff
	case 4:
		// autodecrement - step depends on which register is in use:
		addressInc = 2
		if (reg < 6) && (accessMode&ByteMode > 0) {
			addressInc = 1
		}
		c.Registers[reg] = (c.Registers[reg] - addressInc) & 0xffff
		virtAddress = c.Registers[reg] & 0xffff
	case 5:
		// autodecrement deferred
		virtAddress = c.mmunit.ReadMemoryWord((c.Registers[reg] - 2) & 0xffff)
	case 6:
		// index mode -> read next word to get the basis for address, add value in Register
		offset := c.Fetch()
		virtAddress = offset + c.Registers[reg]

		// TODO: really not needed?
		// c.Registers[7] = (c.Registers[7] + 2) & 0xffff
	case 7:
		baseAddr := c.mmunit.ReadMemoryWord(c.Registers[7])
		virtAddress = (baseAddr + c.Registers[reg]) & 0xffff
		virtAddress = c.mmunit.ReadMemoryWord(virtAddress)
		// increment program counter register
		c.Registers[7] = (c.Registers[7] + 2) & 0xffff
	}
	// all-catcher return
	return virtAddress, nil
}

// Push to processor stack
func (c *CPU) Push(v uint16) {
	c.Registers[6] -= 2
	c.mmunit.WriteMemoryWord(c.Registers[6], v)
}

// Pop from CPU stack
func (c *CPU) Pop() uint16 {
	val := c.mmunit.ReadMemoryWord(c.Registers[6])
	c.Registers[6] += 2
	return val
}

// Reset CPU
// TODO: finish implementation
func (c *CPU) Reset() {
	for i := 0; i < 7; i++ {
		c.Registers[i] = 0
	}
	for i := 0; i < 16; i++ {
		c.mmunit.PAR[i] = 0
		c.mmunit.PDR[i] = 0
	}
	c.mmunit.SR0 = 0
	c.ClockCounter = 0
	c.mmunit.unibus.Rk01.Reset()
	c.State = CPURUN
}
