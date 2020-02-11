package mmu

import (
	"fmt"
	"pdp/psw"
)

// memory related constans (by far not all needed -- figuring out as while writing)
const (
	// ByteMode -> Read addresses by byte, not by word (?)
	ByteMode = 1

	//IO base UNIBUS adresses:
	IObaseVirtual = 0160000
	IObase18bit   = 0760000
	IObaseUnibus  = 017000000
	IObase22bit   = 017760000
)

// MMU related functionality - translating virtual to physical addresses.
type MMU struct {
	Memory          *[4 * 1024 * 1024]byte
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
	// relying on zero-initialization
	// 0 = kernel
	// 1 = super
	// 2 = unused
	// 3 = user
	MMUPar [4][16]int16

	// Page description Regiserts
	MMUPDR [4][16]int16

	// Processor status word.
	// MMU puts it in either 0177776 or 0777776 location
	// (and I'm not really sure which is user where)
	Psw psw.PSW
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

	// TODO: is it used anywhere?
	// pdr: address pointed by PAR:
	pdr := m.MMUPDR[m.MMUMode][page]
	fmt.Printf("DEBUG: pdr: %#o\n", pdr)

	// physical address calculation:
	physicalAddress = (uint32(m.MMUPar[m.MMUMode][page]<<6) + uint32(virtualAddress&0x1FFF)) & 0x3fffff
	return physicalAddress
}

// ReadMemoryWord reads 16 bit word from the memory
// params:
// addr : 16 bit virtual address
// returns: 16 bit word
func (m *MMU) ReadMemoryWord(addr uint16) uint16 {
	// NO MMU SUPPORT SO FAR, 16 bit ADDRESSING ONLY HERE!!
	lowerBit := m.Memory[addr]
	higherBit := m.Memory[addr+1]
	returnWord := uint16(higherBit) << 8
	return returnWord | uint16(lowerBit)
}

// ReadMemoryByte returns single byte from the memory
// TODO: add virtual to physical address translation call
// params:
// addr: 16 bit virtual address
// returns: byte
func (m *MMU) ReadMemoryByte(addr uint16) byte {
	return 0
}

// WriteMemoryWord writes 16 bit word at the memory address
// passed as function parameter
// params:
// addr: 16 bit virtual address
// data: 16 bit word to write
// returns: error
// TODO: is it proper order?
func (m *MMU) WriteMemoryWord(addr, data uint16) error {
	lowerByte := byte(data & 0xff)
	upperByte := byte(data >> 8)
	m.Memory[addr] = lowerByte
	m.Memory[addr+1] = upperByte
	return nil
}

// WriteMemoryByte writes 8 bit byte at the memory address
// passed as function parameter
// params:
// addr: 16 bit virtual addr
// data: byte to be written
// returns: error
func (m *MMU) WriteMemoryByte(addr uint16, data byte) error {
	return nil
}
