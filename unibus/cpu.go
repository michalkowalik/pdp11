package unibus

import (
	"fmt"
	"pdp/interrupts"
	"strings"
)

// memory related constants (by far not all needed -- figuring out as  writing)
const (
	// CPU state: Run / Halt / Wait:
	HALT   = 0
	CPURUN = 1
	WAIT   = 2

	// KernelMode - kernel cpu mode const
	KernelMode = 0
	// UserMode - user cpu mode const
	UserMode = 3
)

// add debug output to the console
var debug = false
var trapDebug = true

// CPU type:
type CPU struct {
	Registers [8]uint16
	State     int

	// system stack pointers: kernel, super, illegal, user
	// super won't be needed for pdp11/40:
	KernelStackPointer uint16
	UserStackPointer   uint16

	// memory access is required:
	mmunit *MMU18Bit

	// ClockCounter
	ClockCounter uint16

	// instructions is a map, where key is the opcode,
	// and value is the function executing it
	// the opcode function should append to the following signature:
	// param: instruction int16
	// return: error -> nil if everything went OK
	singleOpOpcodes       map[uint16]func(uint16)
	doubleOpOpcodes       map[uint16]func(uint16)
	rddOpOpcodes          map[uint16]func(uint16)
	controlOpcodes        map[uint16]func(uint16)
	singleRegisterOpcodes map[uint16]func(uint16)
	otherOpcodes          map[uint16]func(uint16)
}

// NewCPU initializes and returns the CPU variable:
func NewCPU(mmunit *MMU18Bit, debugMode bool) *CPU {

	c := CPU{}
	c.mmunit = mmunit
	c.ClockCounter = 0
	debug = debugMode

	// single operand
	c.singleOpOpcodes = make(map[uint16]func(uint16))
	c.doubleOpOpcodes = make(map[uint16]func(uint16))
	c.rddOpOpcodes = make(map[uint16]func(uint16))
	c.controlOpcodes = make(map[uint16]func(uint16))
	c.otherOpcodes = make(map[uint16]func(uint16))
	c.singleRegisterOpcodes = make(map[uint16]func(uint16))

	// single opearnd:
	c.singleOpOpcodes[0100] = c.jmpOp
	c.singleOpOpcodes[0300] = c.swabOp
	c.singleOpOpcodes[05000] = c.clrOp
	c.singleOpOpcodes[0105000] = c.clrOp
	c.singleOpOpcodes[05100] = c.comOp
	c.singleOpOpcodes[0105100] = c.combOp
	c.singleOpOpcodes[05200] = c.incOp
	c.singleOpOpcodes[0105200] = c.incbOp
	c.singleOpOpcodes[05300] = c.decOp
	c.singleOpOpcodes[0105300] = c.decbOp
	c.singleOpOpcodes[05400] = c.negOp
	c.singleOpOpcodes[0105400] = c.negbOp
	c.singleOpOpcodes[05500] = c.adcOp
	c.singleOpOpcodes[0105500] = c.adcOp
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
	c.doubleOpOpcodes[0120000] = c.cmpOp
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
	c.Registers[7] += 2
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

	// 2 operand instruction in RDD format
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

	// haltOp has opcode of 0, easiest to treat it separately
	if instr == 0 {
		return c.otherOpcodes[0]
	}

	// at this point it can be only an invalid instruction:
	fmt.Printf(c.printState(instr))
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
// readWord returns value specified by source or destination part of the operand.
func (c *CPU) readWord(op uint16) uint16 {
	addr := c.GetVirtualByMode(op, 0)
	return c.mmunit.ReadMemoryWord(addr)
}

// read byte
func (c *CPU) readByte(op uint16) byte {
	addr := c.GetVirtualByMode(op, 1)
	return c.mmunit.ReadMemoryByte(addr)
}

// writeWord writes word value into specified memory address
func (c *CPU) writeWord(op, value uint16) {
	addr := c.GetVirtualByMode(op, 0)
	c.mmunit.WriteMemoryWord(addr, value)
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
	// registers
	out := fmt.Sprintf("%s\n", c.DumpRegisters())

	// flags
	out += fmt.Sprintf("%s ", c.mmunit.unibus.psw.GetFlags())

	// instruction
	out += fmt.Sprintf(" instr %06o: %06o   ", c.Registers[7]-2, instruction)

	return out
}

// SetFlag sets CPU carry flag in Processor Status Word
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

// GetFlag returns carry flag
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
	// save processor stack pointers:
	if c.mmunit.Psw.GetMode() == 3 {
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
	c.mmunit.Psw.SwitchMode(m)
}

// Trap handles all Trap / abort events.
func (c *CPU) Trap(trap interrupts.Trap) {
	if debug || trapDebug {
		fmt.Printf("TRAP %o occured: %s\n", trap.Vector, trap.Msg)
	}
	prevPSW := c.mmunit.Psw.Get()

	defer func(vec, prevPSW uint16) {
		t := recover()
		switch t := t.(type) {
		case interrupts.Trap:
			fmt.Printf("RED STACK TRAP!")
			c.mmunit.Memory[0] = c.Registers[7]
			c.mmunit.Memory[1] = prevPSW
			vec = 4
			panic("FATAL")
		case nil:
			break
		default:
			panic(t)
		}
		c.Registers[7] = c.mmunit.ReadWordByPhysicalAddress(uint32(vec))
		c.mmunit.Psw.Set(c.mmunit.ReadWordByPhysicalAddress(uint32(vec) + 2))
		if prevPSW>>14 == 3 {
			c.mmunit.Psw.Set(c.mmunit.Psw.Get() | (1 << 13) | (1 << 12))
		}
	}(trap.Vector, prevPSW)

	if trap.Vector&1 == 1 {
		panic("Trap called with odd vector number!")
	}

	c.SwitchMode(KernelMode)
	c.Push(prevPSW)
	c.Push(c.Registers[7])
}

// GetVirtualByMode returns virtual address extracted from the CPU instruction
// access mode: 0 for Word, 1 for Byte
func (c *CPU) GetVirtualByMode(instruction, accessMode uint16) uint16 {
	addressInc := uint16(2)
	reg := instruction & 7
	addressMode := (instruction >> 3) & 7
	var virtAddress uint16

	// byte mode
	if accessMode == 1 && reg < 6 {
		addressInc = 1
	}

	switch addressMode {
	case 0:
		virtAddress = 0177700 | reg
	case 1:
		// register keeps the address of the address:
		virtAddress = c.Registers[reg]
	case 2:
		// register keeps the address. Increment the value by 2 (word!)
		virtAddress = c.Registers[reg]
		c.Registers[reg] = c.Registers[reg] + addressInc
	case 3:
		// autoincrement deferred --> it doesn't look like byte mode applies here?
		virtAddress = c.mmunit.ReadMemoryWord(c.Registers[reg])
		c.Registers[reg] = c.Registers[reg] + 2
	case 4:
		// autodecrement - step depends on which register is in use:
		c.Registers[reg] = c.Registers[reg] - addressInc
		virtAddress = c.Registers[reg]
	case 5:
		// autodecrement deferred
		c.Registers[reg] = c.Registers[reg] - 2
		virtAddress = c.mmunit.ReadMemoryWord(c.Registers[reg])
	case 6:
		// index mode -> read next word to get the basis for address, add value in Register
		offset := c.Fetch()
		virtAddress = offset + c.Registers[reg]
	case 7:
		offset := c.Fetch()
		virtAddress = c.mmunit.ReadMemoryWord(offset + c.Registers[reg])
	}
	// all-catcher return
	return virtAddress
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
func (c *CPU) Reset() {
	for i := 0; i < 7; i++ {
		c.Registers[i] = 0
	}
	for i := 0; i < 16; i++ {
		c.mmunit.PAR[i] = 0
		c.mmunit.PDR[i] = 0
	}

	c.KernelStackPointer = 0
	c.UserStackPointer = 0
	c.mmunit.SR0 = 0
	c.ClockCounter = 0
	c.mmunit.unibus.Rk01.Reset()
	c.State = CPURUN
}
