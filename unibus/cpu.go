package unibus

import (
	"fmt"
	"log"
	"pdp/interrupts"
	"strings"
)

type CpuState int

const (
	HALT CpuState = iota
	CPURUN
	WAIT
)

// KernelMode - kernel cpu mode const
const KernelMode = 0

// UserMode - user cpu mode const
const UserMode = 3

// add debug output to the console
var (
	debug      = false
	debugQueue *DebugQueue
)

// CPU type:
type CPU struct {
	Registers [8]uint16
	State     CpuState

	KernelStackPointer, UserStackPointer uint16

	unibus *Unibus
	mmunit MMU
	log    *log.Logger

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
func NewCPU(mmunit MMU, unibus *Unibus, debugMode bool, log *log.Logger) *CPU {
	c := CPU{}
	c.mmunit = mmunit
	debug = debugMode
	c.unibus = unibus
	c.log = log

	if debug {
		debugQueue = NewQueue(1000)
	}

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
	c.controlOpcodes[0102000] = c.bvcOp
	c.controlOpcodes[0102400] = c.bvsOp
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

	// single register and condition code opcodes
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
	physicalAddress := c.mmunit.Decode(c.Registers[7], false, c.IsUserMode())
	instruction := c.unibus.ReadIO(physicalAddress)

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
	fmt.Printf("%s\n", c.printState(instr))
	if debug {
		for !debugQueue.IsEmpty() {
			i, e := debugQueue.Dequeue()
			if e != nil {
				_ = fmt.Errorf(e.Error())
			}
			fmt.Printf("%s\n", i)
		}

	}
	panic(interrupts.Trap{Vector: interrupts.IntINVAL, Msg: "Invalid Instruction"})
}

// Execute decoded instruction
func (c *CPU) Execute() {
	if c.State == WAIT {
		select {
		case v, ok := <-c.unibus.KeyboardInput:
			if ok {
				c.unibus.TermEmulator.AddChar(v)
			}
		default:
		}
		return
	}

	instruction := c.Fetch()
	if debug {
		debugQueue.Enqueue(fmt.Sprintf("%s %s\n", c.printState(instruction), c.unibus.Disasm(instruction)))
	}

	opcode := c.Decode(instruction)
	opcode(instruction)
}

func (c *CPU) IsUserMode() bool {
	return c.unibus.Psw.GetMode() == UserMode
}

func (c *CPU) IsPrevModeUser() bool {
	return c.unibus.Psw.GetPreviousMode() == UserMode
}

// readWord returns value specified by source or destination part of the operand.
func (c *CPU) readWord(op uint16) uint16 {
	addr := c.GetVirtualAddress(op, 0)
	return c.mmunit.ReadMemoryWord(addr)
}

// read byte
func (c *CPU) readByte(op uint16) byte {
	addr := c.GetVirtualAddress(op, 1)
	return c.mmunit.ReadMemoryByte(addr)
}

// writeWord writes word value into the specified memory address
func (c *CPU) writeWord(op, value uint16) {
	addr := c.GetVirtualAddress(op, 0)
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
	out += fmt.Sprintf("%s ", c.unibus.Psw.GetFlags())

	// instruction
	out += fmt.Sprintf(" instr %06o: %06o   ", c.Registers[7]-2, instruction)

	return out
}

// SetFlag sets CPU carry flag in Processor Status Word
func (c *CPU) SetFlag(flag string, set bool) {
	switch flag {
	case "C":
		c.unibus.Psw.SetC(set)
	case "V":
		c.unibus.Psw.SetV(set)
	case "Z":
		c.unibus.Psw.SetZ(set)
	case "N":
		c.unibus.Psw.SetN(set)
	case "T":
		c.unibus.Psw.SetT(set)
	}
}

// GetFlag returns carry flag
func (c *CPU) GetFlag(flag string) bool {
	switch flag {
	case "C":
		return c.unibus.Psw.C()
	case "V":
		return c.unibus.Psw.V()
	case "Z":
		return c.unibus.Psw.Z()
	case "N":
		return c.unibus.Psw.N()
	case "T":
		return c.unibus.Psw.T()
	}
	return false
}

// SwitchMode switches the kernel / user mode:
func (c *CPU) SwitchMode(mode uint16) {
	previousMode := c.unibus.Psw.GetMode()

	// save processor stack pointers:
	if previousMode == UserMode {
		c.UserStackPointer = c.Registers[6]
	} else {
		c.KernelStackPointer = c.Registers[6]
	}

	// set processor stack:
	if mode == UserMode {
		c.Registers[6] = c.UserStackPointer
	} else {
		c.Registers[6] = c.KernelStackPointer
	}
	*c.unibus.Psw &= 000777
	if mode == UserMode {
		*c.unibus.Psw |= (1 << 15) | (1 << 14)
	}
	if previousMode == UserMode {
		*c.unibus.Psw |= (1 << 13) | (1 << 12)
	}
}

func (c *CPU) GetVirtualAddress(instruction, accessMode uint16) uint16 {
	addressInc := uint16(2)
	reg := instruction & 7
	addressMode := (instruction >> 3) & 7
	var virtAddress uint16

	// byte mode does not apply to the SP and PC
	if accessMode == 1 && reg < 6 {
		addressInc = 1
	}

	switch addressMode {
	case 0:
		// register contains operand
		virtAddress = 0177700 | reg
	case 1:
		// register contains the address of the operand
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
		c.mmunit.SetPage(i, page{par: 0, pdr: 0})
	}

	c.KernelStackPointer = 0
	c.UserStackPointer = 0
	c.mmunit.SetSR0(0)
	c.unibus.Rk01.Reset()
	c.State = CPURUN
}

/*
// debug:
// true if all registers have matching value. don't panic immediately, there might be a panic counter somewhere.
func (c *CPU) timeToDie(registers []uint16) bool {
	for i, v := range c.Registers {
		if registers[i] != v {
			return false
		}
	}
	return true
}
*/
