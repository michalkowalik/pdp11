package cpu

import "fmt"

// CPU type:
type CPU struct {
	registers                   [8]uint16
	statusFlags                 byte
	floatingPointStatusRegister byte
	statusRegister              uint16
}

// cpu should be able to fetch, decode and execute:

// Fetch loads next command from memory
func (c CPU) Fetch() {
	fmt.Printf("CPU Fetch\n")
}

//Decode - missing comment
func (c CPU) Decode() {
	fmt.Printf("Decode..\n")
}

// Execute fetched order
func (c CPU) Execute() {
	fmt.Printf("Execute.. \n")
}

// helper functions:

// DumpRegisters displays register values
func (c CPU) DumpRegisters() {
	fmt.Printf("nothing to see yet \n")
}
