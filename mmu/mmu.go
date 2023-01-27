package mmu

/*

Interfaces and type definitions for the 18 bit MMU, as used in PDP11/40

*/

type uint18 uint32

type mmu interface {
	Read16(addr uint18) uint16
	Write16(addr uint18, value uint16)

	MmuEnabled() bool

	// let's add decode to the interface, will make the testing easier
	Decode(a uint16, w, user bool) uint18
}
