package cpu

import "fmt"

// CPU elements:
type cpu struct {
	registers                   [8]uint16
	statusFlags                 byte
	floatingPointStatusRegister byte
	statusRegister              uint16
}

// cpu should be able to fetch, decode and execute:

// Fetch loads next command from memory
func (c cpu) Fetch() {
	fmt.Printf("CPU Fetch\n")
}

//Decode - missing comment
func (c cpu) Decode() {
	fmt.Printf("Decode..\n")
}

func (c cpu) Execute() {
	fmt.Printf("Execute.. \n")
}

// helper functions:

// DumpRegisters displays register values
func (c cpu) DumpRegisters() {
	fmt.Printf("nothing to see yet \n")
}
