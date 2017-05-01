package main

import (
	"fmt"
	"pdp/system"
)

func main() {
	fmt.Printf("Starting PDP-11/70 emulator..\n")
	pdp := system.InitializeSystem()
	pdp.Noop()
}
