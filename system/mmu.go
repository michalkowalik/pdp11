package system

import (
	"fmt"
)

// PDP11/70 can be equipped with up to 4MB of RAM.
// As the memory addresses suppored by the cpu are 16bit long,
// that means, there's only 64K of memory directly accessible to the program.
// To circumvent it, the MMU (Memory Management Unit) has to be used.
// abbreviation used below:
// MMR - Memory Management Register
// PAR - Page Address Register
// PAF - Page Address Field
// APF - Active Page Field

// on top of that, pdp 11/(44,70) are using Instruction and data memory pages
// hence the I/D marker on the virtual address.  --> Hence both MMUPar and MMUPDR

// On a real PDP 11 the memory registers are located in thee uppermost 8K of RAM address space
// along with the Unibus I/O device registers.

// MMR composition:
// 15 | 14 | 13 | 12 | 11 | 10 | 9 | 8 | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 MMR0
//nonr leng read trap unus unus ena mnt cmp  -mode- i/d  --page--   enable

// memory related constans (by far not all needed -- figuring out as while writing)

// ByteMode -> Read addresses by byte, not by word (?)
const ByteMode = 1

// ReadMode -> Read from main memory
const ReadMode = 2

// WriteMode -> Write from main memory
const WriteMode = 4

// ModifyWord ->  Read and write word in memory
const ModifyWord = ReadMode | WriteMode

// ModifyByte -> Read and write byte in memory
const ModifyByte = ReadMode | WriteMode | ByteMode

// MMU related functionality - translating virtual to physical addresses.
type MMU struct {
	MMR             [4]int16 // Memory Management Registers
	MMR3Map         [4]int16 // Map from mode to MMR3 I/D bit mask
	MMUEnable       int16
	MMULastMode     int
	MMMULastVirtual uint16

	// Current memory management mode:
	// 0 = kernel
	// 1 = super
	// 2 = undefined
	// 3 = user
	// Typically this should be set by writing PSW (Processor Status Word).
	// There are though few instructions moving data between address spaces.
	// Those modify the MMU Mode without writing the PSW and set it back if all worked OK
	MMUMode int

	// MMU Page Address Registers
	// TODO -> Why 16 per map? Shouldn't be 8?
	// relying on zero-initialization
	// 0 = kernel
	// 1 = super
	// 2 = unused
	// 3 = user
	MMUPar [4][16]int16

	// memory managemnt PDR registers by mode
	MMUPRD [4][16]int16
}

// MapVirtualToPhysical maps the 17 bit I/D virtual address to a 22 bit physical address
// TODO: All checks and specifics (i.e. 18bit calculation)
func (m *MMU) MapVirtualToPhysical(virtualAddress uint16, accessMask int16) uint32 {
	var physicalAddress uint32
	// this access doesn't require MMU
	if (accessMask & m.MMUEnable) == 0 {
		physicalAddress = uint32(virtualAddress & 0xffff) // virtual addr. without mmu is 16 bit!
		m.MMMULastVirtual = virtualAddress & 0xffff
		// TODO: add boundary checks, throw trap if fail
		return physicalAddress
		// in any other case, MMU is required:
	}

	// In any other case MMU translation is required.
	m.MMMULastVirtual = virtualAddress

	// address page -> oldest 3 bits points to the MMR keeping it
	page := virtualAddress >> 13
	fmt.Printf("DEBUG: page: %#o\n", page)
	// TODO: Add error checking like in pdp11.js, line 484

	// pdr: address pointed by PAR:
	pdr := m.MMUPRD[m.MMUMode][page]
	fmt.Printf("DEBUG: pdr: %#o\n", pdr)

	// physical address calculation:
	physicalAddress = (uint32(m.MMUPar[m.MMUMode][page]<<6) + uint32(virtualAddress&0x1FFF)) & 0x3fffff
	return physicalAddress
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
func (m *MMU) GetVirtualByMode(instruction, accessMode uint16) uint32 {
	// var addressInc int
	// reg := instruction & 7
	addressMode := (instruction >> 3) & 7
	switch addressMode {
	case 0:
		return 0
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 3
	case 4:
		return 4
	case 5:
		return 5
	case 6:
		return 6
	case 7:
		return 7
	}
	// deal with change of address pointer in the register if needed:

	// all-catcher return
	return 0
}

// ReadMemoryWord reads 16 bit word from the memory
// params:
// addr : 16 bit virtual address
// returns: 16 bit word
func (m *MMU) ReadMemoryWord(addr int16) int16 {
	return 0
}

// ReadMemoryByte returns single byte from the memory
// params:
// addr: 16 bit virtual address
// returns: byte
func (m *MMU) ReadMemoryByte(addr int16) byte {
	return 0
}

// WriteMemoryWord writes 16 bit word at the memory address
// passed as function parameter
// params:
// addr: 16 bit virtual address
// data: 16 bit word to write
// returns: error
func (m *MMU) WriteMemoryWord(addr, data int16) error {
	return nil
}

// WriteMemoryByte writes 8 bit byte at the memory address
// passed as function parameter
// params:
// addr: 16 bit virtual addr
// data: byte to be written
// returns: error
func (m *MMU) WriteMemoryByte(addr int16, data byte) error {
	return nil
}
