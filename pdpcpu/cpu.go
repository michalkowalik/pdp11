package pdpcpu

import "fmt"

// CPU type:
type CPU struct {
	registers                   [8]uint16
	statusFlags                 byte
	floatingPointStatusRegister byte
	statusRegister              uint16
}

// cpu should be able to fetch, decode and execute:

// Fetch next instruction from memory
func (c CPU) Fetch() {
	fmt.Printf("CPU Fetch\n")
}

//Decode fetched instruction
func (c CPU) Decode() {
	fmt.Printf("Decode..\n")
}

// Execute decoded instruction
func (c CPU) Execute() {
	fmt.Printf("Execute.. \n")
}

// helper functions:

// DumpRegisters displays register values
func (c CPU) DumpRegisters() {
	for i, reg := range c.registers {
		fmt.Printf("R%d: 0x%x\n", i, reg)
	}
}
