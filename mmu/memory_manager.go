package mmu

// MemoryManager is an interface for different mmu implementation.
// Initially 2 implementations are available, for 18 and 22 bit virtual memory size.MemoryManager
type MemoryManager interface {

	// ReadMemoryWord returns content of the physical address mapped
	// through 16 bit virtual "addr" address
	ReadMemoryWord(addr uint16) uint16

	// ReadMemoryByte returns byte content of the physical address mapped
	// through the 16 bit virtual "addr" address
	ReadMemoryByte(addr uint16) byte

	// WriteMemoryWord writes "data" to virtual address "addr"
	WriteMemoryWord(addr, data uint16) error

	// WriteMemoryByte writes to address "addr" content of the "data" byte
	WriteMemoryByte(addr uint16, data byte) error
}
