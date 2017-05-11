package system

import (
	"errors"
	"fmt"
	"pdp/pdpcpu"

	"github.com/jroimartin/gocui"
)

// System definition.
type System struct {
	Memory [4 * 1024 * 1024]byte
	CPU    *pdpcpu.CPU

	mmuEnabled bool

	// Unibus map registers
	UnibusMap [32]int16

	// console and status output:
	statusView  *gocui.View
	consoleView *gocui.View
	regView     *gocui.View
}

var mmu MMU

// InitializeSystem initializes the emulated PDP-11/44 hardware
func InitializeSystem(statusView, consoleView, regView *gocui.View) *System {
	sys := new(System)
	sys.CPU = pdpcpu.New()
	sys.statusView = statusView
	sys.consoleView = consoleView
	sys.regView = regView

	// start emulation with disabled mmu:
	sys.mmuEnabled = false

	// point mmu to memory:
	mmu = MMU{}
	mmu.Memory = &sys.Memory

	fmt.Fprintf(statusView, "Initializing PDP11 CPU...\n")

	return sys
}

// Noop is a dummy function just to keep go compiler happy for a while
func (sys *System) Noop() {
	fmt.Fprintf(sys.consoleView, ".. Noop ..\n")
}

// GetVirtualByMode maps six bit instruction operand to a 17 bit virtual address space.
// Below follows a copied comment from  http://skn.noip.me/pdp11/pdp11.html
// Instruction operands are six bits in length - three bits for the mode and three
// for the register. The 17th I/D bit in the resulting virtual address represents
// whether the reference is to Instruction space or Data space - which depends on
// combination of the mode and whether the register is the Program Counter (register 7).
//
// The eight modes are:-
//              0       R               no valid virtual address
//              1       (R)             operand from I/D depending if R = 7
//              2       (R)+            operand from I/D depending if R = 7
//              3       @(R)+           address from I/D depending if R = 7 and operand from D space
//              4       -(R)            operand from I/D depending if R = 7
//              5       @-(R)           address from I/D depending if R = 7 and operand from D space
//              6       x(R)            x from I space but operand from D space
//              7       @x(R)           x from I space but address and operand from D space
//
// Stack limit checks are implemented for modes 1, 2, 4 & 6 (!)
//
// Also need to keep CPU.MMR1 updated as this stores which registers have been
// incremented and decremented so that the OS can reset and restart an instruction
// if a page fault occurs.
// accessMode -> one of the Read, write, modify -> binary value, 0, 1, 2, 4 etc.
// TODO: Move to CPU module?
func (sys *System) GetVirtualByMode(instruction, accessMode uint16) (uint32, error) {

	c := sys.CPU

	var addressInc int16
	reg := instruction & 7
	addressMode := (instruction >> 3) & 7
	var virtAddress uint32

	switch addressMode {
	case 0:
		return 0, errors.New("Wrong address mode - throw trap?")
	case 1:
		// register keeps the address:
		virtAddress = uint32(c.Registers[reg])
	case 2:
		// register keeps the address. Increment the value by 2 (word!)
		addressInc = 2
		virtAddress = uint32(c.Registers[reg])
	case 3:
		// autoincrement deferred
		// TODO: ADD special cases (R6 and R7)
		addressInc = 2
		virtAddress = uint32(c.Registers[reg])
	case 4:
		// autodecrement - step depends on which register is in use:
		addressInc = -2
		if (reg < 6) && (accessMode&ByteMode > 0) {
			addressInc = -1
		}
		virtAddress = uint32(int32(c.Registers[reg])+int32(addressInc)) & 0xffff
		addressInc = 0
	case 5:
		// autodecrement deferred
		virtAddress = uint32(c.Registers[reg]-2) & 0xffff
	case 6:
		// index mode -> read next word to get the basis for address, add value in Register
		baseAddr := mmu.ReadMemoryWord(c.Registers[7])
		virtAddress = uint32((baseAddr + c.Registers[reg]) & 0xffff)

		// increment program counter register
		c.Registers[7] = (c.Registers[7] + 2) & 0xffff
	case 7:
		baseAddr := mmu.ReadMemoryWord(c.Registers[7])
		virtAddress = uint32((baseAddr + c.Registers[reg]) & 0xffff)

		// increment program counter register
		c.Registers[7] = (c.Registers[7] + 2) & 0xffff
	}
	// deal with change of address pointer in the register if needed:

	// all-catcher return
	return virtAddress, nil
}
