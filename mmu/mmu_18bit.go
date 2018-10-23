package mmu

import (
	"fmt"
	"pdp/interrupts"
	"pdp/psw"
	"pdp/unibus"
)

// MMU18Bit implements the 18 bit memory management unit
// max 256KB RAM (128K words), enough to run early unix versions
type MMU18Bit struct {

	// Memory : Physical memory
	Memory [128 * 1024]uint16

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
	Psw *psw.PSW

	// it's also convenient to keep a pointer to Unibus..
	unibus *unibus.Unibus
}

// MaxMemory - available for user, 248k
const MaxMemory = 0760000

// MaxTotalMemory - highest address available. Top  4k used by unibus
const MaxTotalMemory = 0777776

// UnibusMemoryBegin in 16 bit mode
const UnibusMemoryBegin = 0170000

// New returns the new MMU18Bit struct
func New(psw *psw.PSW, unibus *unibus.Unibus) *MMU18Bit {
	mmu := MMU18Bit{}
	mmu.Psw = psw
	mmu.unibus = unibus
	return &mmu
}

// return true if MMU enabled - controlled by 1 bit in the SR0
func (m *MMU18Bit) mmuEnabled() bool {
	return m.SR0&1 == 1
}

// mapVirtualToPhysical retuns physical 18 bit address for the 16 bit virtual
func (m *MMU18Bit) mapVirtualToPhysical(virtualAddress uint16) uint32 {
	if !m.mmuEnabled() {
		addr := uint32(virtualAddress)
		if addr >= UnibusMemoryBegin {
			addr += 0600000
		}
		return addr
	}

	// if bits 14 and 15 in PSW are set -> system in kernel mode
	currentUser := uint16(0)
	if m.Psw.GetMode() > 0 {
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
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Reading from odd address: %o", physicalAddress)})
	}
	if physicalAddress > MaxTotalMemory {
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Read from invalid address")})
	}
	if physicalAddress < MaxMemory {
		return m.Memory[physicalAddress>>1]
	}

	// IO Page:
	data, err := m.unibus.ReadIOPage(physicalAddress, false)
	if err != nil {
		panic(interrupts.Trap{
			Vector: interrupts.INTFault,
			Msg:    "mmu.go, ReadMemoryWord"})
	}
	return data
}

// ReadMemoryByte reads a byte from virtual address addr
func (m *MMU18Bit) ReadMemoryByte(addr uint16) byte {
	defer func() {
		t := recover()
		switch t := t.(type) {
		case interrupts.Trap:
			m.unibus.Traps <- t
		case nil:
			// ignore
		default:
			panic(t)
		}
	}()

	// Zero the lowest byte, to avoid reading a word from odd address
	val := m.ReadMemoryWord(addr & 0xFFFE)
	if addr&1 > 0 {
		return byte(val >> 8)
	}
	return byte(val & 0xFF)
}

// WriteMemoryWord writes a word to the location pointed by virtual address addr
func (m *MMU18Bit) WriteMemoryWord(addr, data uint16) {
	physicalAddress := m.mapVirtualToPhysical(addr)
	if (physicalAddress & 1) == 1 {
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    "Write to odd address"})
	}
	if physicalAddress < MaxMemory {
		m.Memory[physicalAddress>>1] = data
	} else if physicalAddress == MaxMemory {
		// update PSW
		*m.Psw = psw.PSW(data)
	} else if physicalAddress >= MaxMemory && physicalAddress <= MaxTotalMemory {
		m.unibus.WriteIOPage(physicalAddress, data, false)
	} else {
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    "Write to invalid address"})
	}
}

// WriteMemoryByte writes a byte to the location pointed by virtual address addr
func (m *MMU18Bit) WriteMemoryByte(addr uint16, data byte) {
	defer func() {
		t := recover()
		switch t := t.(type) {
		case interrupts.Trap:
			m.unibus.Traps <- t
		case nil:
			// ignore
		default:
			panic(t)
		}
	}()

	physicalAddress := m.mapVirtualToPhysical(addr)
	var wordData uint16
	if addr&1 == 0 {
		wordData = (m.Memory[physicalAddress>>1] & 0xFF00) | uint16(data)
	} else {
		wordData = (m.Memory[physicalAddress>>1] & 0xFF) | (uint16(data) << 8)
	}
	m.WriteMemoryWord(addr, wordData)
}
