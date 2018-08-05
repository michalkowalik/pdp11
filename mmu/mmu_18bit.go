package mmu

import (
	"pdp/interrupts"
	"pdp/psw"
	"pdp/unibus"
)

// MMU18Bit implements the 18 bit memory management unit
// max 256KB RAM (128K words), enough to run early unix versions
type MMU18Bit struct {

	// Memory : Physical memory
	Memory *[128 * 1024]uint16

	// PAR : Page Address Registers
	PAR [16]uint16

	// PDR : Page Description Registers
	PDR [16]uint16

	// APR : Active Page Register - 8 of them
	APR [8]uint16

	// MMU Status Register 0 (Status and error indicator 0)
	SR0 uint16

	// MMU Status Register 2 (Status and erorr indicator 2)
	SR2 uint16

	// it's convenient to have a pointer to processor status word available all the time
	psw *psw.PSW

	// it's also convenient to keep a pointer to Unibus..
	unibus *unibus.Unibus
}

// MaxMemory - available for user, 248k
const MaxMemory = 0760000

// MaxTotalMemory - highest address available. Top  4k used by unibus
const MaxTotalMemory = 0777777

// New returns the new MMU18Bit struct
func New(psw *psw.PSW, unibus *unibus.Unibus) *MMU18Bit {
	mmu := MMU18Bit{}
	mmu.psw = psw
	mmu.unibus = unibus
	return &mmu
}

// mapVirtualToPhysical retuns physical 18 bit address for the 16 bit virtual
func (m *MMU18Bit) mapVirtualToPhysical(
	virtualAddress uint16) uint32 {
	// if bits 14 and 15 in PSW are set -> system in kernel mode
	currentUser := uint16(0)
	if (m.psw.Get() >> 14) > 0 {
		currentUser += 8
	}

	currentPAR := m.PAR[(virtualAddress>>13)+currentUser]
	block := (virtualAddress >> 6) & 0177
	displacement := virtualAddress & 077
	return uint32(uint32((block+currentPAR)<<6+displacement) & 0777777)
}

// ReadMemoryWord reads a word from virtual address addr
func (m *MMU18Bit) ReadMemoryWord(addr uint16) uint16 {
	physicalAddress := m.mapVirtualToPhysical(addr)
	if (physicalAddress & 1) == 1 {
		m.unibus.SendTrap(interrupts.INTBus)
		return 0
	}
	if physicalAddress < MaxMemory {
		return m.Memory[physicalAddress>>1]
	}
	if physicalAddress >= MaxMemory && physicalAddress <= MaxTotalMemory {
		data, err := m.unibus.ReadIOPage(physicalAddress, false)
		if err != nil {
			m.unibus.SendTrap(interrupts.INTFault)
			return 0
		}
		return data
	}

	// if everything else fails:
	m.unibus.SendTrap(interrupts.INTBus)
	return 0
}

// ReadMemoryByte reads a byte from virtual address addr
func (m *MMU18Bit) ReadMemoryByte(addr uint16) byte {
	return 0
}

// WriteMemoryWord writes a word to the location pointed by virtual address addr
func (m *MMU18Bit) WriteMemoryWord(addr, data uint16) error {
	return nil
}

// WriteMemoryByte writes a byte to the location pointed by virtual address addr
func (m *MMU18Bit) WriteMemoryByte(addr, data byte) error {
	return nil
}
