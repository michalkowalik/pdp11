package pdpcpu

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

// CPU type:
type CPU struct {
	Registers                   [8]uint16
	statusFlags                 byte
	floatingPointStatusRegister byte
	statusRegister              uint16

	// instructions is a map, where key is the opcode,
	// and value is the function executing it
	// the opcode function should append to the following signature:
	// param: instruction int16
	// return: error -> nil if everything went OK
	opcodes map[int16](func(int16) error)
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
func New() *CPU {
	c := CPU{}

	// single operand
	c.opcodes = make(map[int16](func(int16) error))
	c.opcodes[050] = clrOp
	c.opcodes[06] = addOp

	return &c
}

// cpu should be able to fetch, decode and execute:

// Fetch next instruction from memory
func (c *CPU) Fetch() {
	fmt.Printf("CPU Fetch\n")
}

//Decode fetched instruction
func (c *CPU) Decode() {
	fmt.Printf("Decode..\n")
}

// Execute decoded instruction
func (c *CPU) Execute() {
	fmt.Printf("Execute.. \n")
}

// helper functions:

// readWord returns value specified by source or destination part of the operand.
func (c *CPU) readWord(op int16) uint16 {
	// check mode:
	mode := op >> 3
	register := op & 07
	switch mode {
	case 0:
		//value directly in register
		return c.Registers[register]
	default:
		return 0
	}
}

// DumpRegisters displays register values
func (c *CPU) DumpRegisters(regView *gocui.View) {
	for i, reg := range c.Registers {
		fmt.Fprintf(regView, " |R%d: %#o | ", i, reg)
	}
}

// single operand cpu instructions:
func clrOp(instruction int16) error {
	return nil
}

// double operand cpu instructions:
func addOp(instruction int16) error {
	return nil
}
