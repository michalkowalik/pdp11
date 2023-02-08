package unibus

/*
Interfaces and type definitions for the MMU
(should definitely work for 18bit MMU, as used in 11/40)
*/

// Uint18 is being unsed only in context of physical 18 bit address.
type Uint18 uint32

type page struct {
	par, pdr uint16
}

type MMU interface {
	Read16(addr Uint18) uint16
	ReadMemoryWord(addr uint16) uint16
	ReadMemoryByte(addr uint16) byte

	Write16(addr Uint18, data uint16)
	WriteMemoryWord(addr, data uint16)
	WriteMemoryByte(addr uint16, data byte)

	MmuEnabled() bool

	// let's add decode to the interface, will make the testing easier
	Decode(a uint16, w, user bool) Uint18

	// SR0 getter and setter
	SetSR0(v uint16)
	SetSR2(v uint16)
	GetSR0() uint16
	GetSR2() uint16

	SetPage(i int, p page)

	// Debugging methods
	DumpMemory() error
}
