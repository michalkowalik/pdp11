package unibus

import (
	"fmt"
	"pdp/interrupts"
	"pdp/psw"
)

type pdr uint16

// MMU18Bit implements the 18 bit memory management unit
// max 256KB RAM (128K words), enough to run early unix versions
type MMU18Bit struct {

	// Memory : Physical memory
	Memory [128 * 1024]uint16

	// PAR : Page Address Registers
	PAR [16]uint16

	// PDR : Page Description Registers
	PDR [16]pdr

	// APR : Active Page Register - 8 of them
	APR [8]uint16

	// MMU Status Register 0 (Status and error indicator 0)
	SR0 uint16

	// MMU Status Register 2 (Status and erorr indicator 2)
	SR2 uint16

	// it's convenient to have a pointer to processor status word available all the time
	Psw *psw.PSW

	// it's also convenient to keep a pointer to Unibus..
	unibus *Unibus

	// MMUDebugMode - enables output of the MMU internals to the terminal
	MMUDebugMode bool
}

// MMUDebugMode - enables output of the MMU internals to the terminal
var MMUDebugMode = false

// MaxMemory - available for user, 248k
const MaxMemory = 0760000

// MaxTotalMemory - highest address available. Top  4k used by unibus
const MaxTotalMemory = 0777776

// UnibusMemoryBegin in 16 bit mode
const UnibusMemoryBegin = 0170000

// RegisterAddressBegin in 18bit space
const RegisterAddressBegin = 0777700

// NewMMU returns the new MMU18Bit struct
func NewMMU(psw *psw.PSW, unibus *Unibus) *MMU18Bit {
	mmu := MMU18Bit{}
	mmu.Psw = psw
	mmu.unibus = unibus
	return &mmu
}

// few helper functions to check the page status with Page description register
func (p pdr) read() bool     { return p&2 == 2 }
func (p pdr) write() bool    { return p&6 == 6 }
func (p pdr) ed() bool       { return p&8 == 8 }
func (p pdr) length() uint16 { return (uint16(p) >> 8) & 0x7F }

// MmuEnabled returns true if MMU enabled - controlled by 1 bit in the SR0
func (m *MMU18Bit) MmuEnabled() bool {
	return m.SR0&1 == 1
}

// Return Page Address or Page Description Register. Register address in virtual address.
func (m *MMU18Bit) readPage(address uint32) uint16 {
	i := ((address & 017) >> 1)

	// kernel space:
	if (address >= 0772300) && (address < 0772320) {
		return uint16(m.PDR[i])
	}
	if (address >= 0772340) && (address < 0772360) {
		return m.PAR[i]
	}

	// user space:
	if (address >= 0777600) && (address < 0777620) {
		return uint16(m.PDR[i+8])
	}
	if (address >= 0777640) && (address < 0777660) {
		return m.PAR[i+8]
	}
	panic(interrupts.Trap{
		Vector: interrupts.INTBus,
		Msg:    fmt.Sprintf("Attempt to read from invalid address %06o", address)})
}

// Modify memory page:
func (m *MMU18Bit) writePage(address uint32, data uint16) {
	i := ((address & 017) >> 1)

	// kernel space:
	if (address >= 0772300) && (address < 0772320) {
		m.PDR[i] = pdr(data)
		return
	}
	if (address >= 0772340) && (address < 0772360) {
		m.PAR[i] = data
		return
	}

	// user space:
	if (address >= 0777600) && (address < 0777620) {
		m.PDR[i+8] = pdr(data)
		return
	}
	if (address >= 0777640) && (address < 0777660) {
		m.PAR[i+8] = data
		return
	}
	panic(interrupts.Trap{
		Vector: interrupts.INTBus,
		Msg:    fmt.Sprintf("Attempt to read from invalid address %06o", address)})
}

// mapVirtualToPhysical retuns physical 18 bit address for the 16 bit virtual
// mode: 0 for kernel, 3 for user
func (m *MMU18Bit) mapVirtualToPhysical(virtualAddress uint16, writeMode bool, mode uint16) uint32 {

	if !m.MmuEnabled() {
		addr := uint32(virtualAddress)
		if addr >= UnibusMemoryBegin {
			addr += 0600000
		}
		return addr
	}

	// if bits 14 and 15 in PSW are set -> system in kernel mode
	currentUser := uint16(0)
	if mode > 0 {
		currentUser += 8
	}
	offset := (virtualAddress >> 13) + currentUser
	if m.MMUDebugMode {
		fmt.Printf("MMU: write mode: %v, offset: %o, PDR[offset]: %o\n", writeMode, offset, m.PDR[offset])
	}

	// check page availability in PDR:
	if writeMode && !m.PDR[offset].write() {
		m.SR0 = (1 << 13) | 1
		m.SR0 |= (virtualAddress >> 12) & ^uint16(1)

		if m.MMUDebugMode {
			fmt.Printf("modified SR0: %o\n", m.SR0)
		}

		// check for user mode
		if m.unibus.psw.GetMode() == 3 {
			m.SR0 |= (1 << 5) | (1 << 6)
		}

		// set SR2 to the value of current program counter
		m.SR2 = m.unibus.PdpCPU.Registers[7]
		panic(interrupts.Trap{
			Vector: interrupts.INTFault,
			Msg:    "Abort: write on read-only page"})
	}

	if !m.PDR[offset].read() {
		m.SR0 = (1 << 15) | 1
		m.SR0 |= (virtualAddress >> 12) & ^uint16(1)

		// check for user mode
		if m.unibus.psw.GetMode() == 3 {
			m.SR0 |= (1 << 5) | (1 << 6)
		}
		m.SR2 = m.unibus.PdpCPU.Registers[7]
		panic(interrupts.Trap{
			Vector: interrupts.INTFault,
			Msg:    "Abort: read on no-access page"})
	}

	currentPAR := m.PAR[offset]
	block := (virtualAddress >> 6) & 0177
	displacement := virtualAddress & 077

	if m.MMUDebugMode {
		fmt.Printf(
			"MMU: block: %o, displacement: %o, currentPAR: %o\n",
			block, displacement, currentPAR)
	}

	// check if page length not exceeded
	if m.PDR[offset].ed() && block < m.PDR[offset].length() ||
		!m.PDR[offset].ed() && block > m.PDR[offset].length() {
		m.SR0 = (1 << 14) | 1
		m.SR0 |= (virtualAddress >> 12) & ^uint16(1)

		// check for user mode
		if m.unibus.psw.GetMode() == 3 {
			m.SR0 |= (1 << 5) | (1 << 6)
		}
		m.SR2 = m.unibus.PdpCPU.Registers[7]
		panic("PAGE LENGTH EXCEEDED")
	}

	// set PDR W byte:
	if writeMode {
		m.PDR[offset] |= 1 << 6
	}

	physAddress := ((uint32(block) + uint32(currentPAR)) << 6) + uint32(displacement)

	if m.MMUDebugMode {
		fmt.Printf(m.dumpRegisters())
	}

	m.MMUDebugMode = false
	return physAddress
}

// ReadMemoryWord reads a word from virtual address addr
// Funny complication: A trap needs to be thrown if case it is an odd address
// but not if it it is an register address -> then it's OK.
func (m *MMU18Bit) ReadMemoryWord(addr uint16) uint16 {
	physicalAddress := m.mapVirtualToPhysical(addr, false, m.Psw.GetMode())
	return m.ReadWordByPhysicalAddress(physicalAddress)
}

// ReadWordByPhysicalAddress - needed to read by physicall address
func (m *MMU18Bit) ReadWordByPhysicalAddress(addr uint32) uint16 {

	if !(addr&RegisterAddressBegin == RegisterAddressBegin) && ((addr & 1) == 1) {
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Reading from odd address: %o", addr)})
	}
	if addr > MaxTotalMemory {
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    fmt.Sprintf("Read from invalid address")})
	}
	if addr < MaxMemory {
		return m.Memory[addr>>1]
	}

	// IO Page:
	data, err := m.unibus.ReadIOPage(addr, false)
	if err != nil {
		panic(interrupts.Trap{
			Vector: interrupts.INTFault,
			Msg:    "mmu.go, ReadMemoryWord"})
	}
	return data
}

// ReadMemoryByte reads a byte from virtual address addr
func (m *MMU18Bit) ReadMemoryByte(addr uint16) byte {
	physicalAddress := m.mapVirtualToPhysical(addr&0xFFFE, false, m.Psw.GetMode())
	memWord := m.ReadWordByPhysicalAddress(physicalAddress)

	if physicalAddress&0777770 == RegisterAddressBegin {
		return byte(m.unibus.PdpCPU.Registers[addr&7] & 0xFF)
	}

	if addr&1 > 0 {
		return byte(memWord >> 8)
	}
	return byte(memWord & 0xFF)
}

// WriteMemoryWord writes a word to the location pointed by virtual address addr
func (m *MMU18Bit) WriteMemoryWord(addr, data uint16) {
	physicalAddress := m.mapVirtualToPhysical(addr, true, m.Psw.GetMode())
	m.WriteWordByPhysicalAddress(physicalAddress, data)
}

// WriteWordByPhysicalAddress - write word to physical address
func (m *MMU18Bit) WriteWordByPhysicalAddress(addr uint32, data uint16) {
	if !(addr&RegisterAddressBegin == RegisterAddressBegin) && ((addr & 1) == 1) {
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    "Write to odd address"})
	}

	if addr < MaxMemory {
		m.Memory[addr>>1] = data
	} else if addr > MaxMemory && addr <= MaxTotalMemory {
		m.unibus.WriteIOPage(addr, data, false)
	} else {
		panic(interrupts.Trap{
			Vector: interrupts.INTBus,
			Msg:    "Write to invalid address"})
	}
}

// WriteMemoryByte writes a byte to the location pointed by virtual address addr
func (m *MMU18Bit) WriteMemoryByte(addr uint16, data byte) {
	// modify register directly:
	if (addr & 0177770) == 0177700 {
		m.unibus.PdpCPU.Registers[addr&7] = uint16(data)
		return
	}

	physicalAddress := m.mapVirtualToPhysical(addr&0xFFFE, false, m.Psw.GetMode())
	wordData := m.ReadWordByPhysicalAddress(physicalAddress)

	if addr&1 == 0 {
		wordData = (wordData & 0xFF00) | uint16(data)
	} else {
		wordData = (wordData & 0xFF) | (uint16(data) << 8)
	}
	physicalAddress = m.mapVirtualToPhysical(addr&0xFFFE, true, m.Psw.GetMode())
	m.WriteWordByPhysicalAddress(physicalAddress, wordData)
}

func (m *MMU18Bit) dumpRegisters() string {
	status := fmt.Sprintf("SR0: %o, SR2: %o\n", m.SR0, m.SR2)
	for i, v := range m.APR {
		status += fmt.Sprintf("APR[%d]: %o\t", i, v)
	}
	status += "\n"

	for i, v := range m.PDR {
		status += fmt.Sprintf("PDR[%d]: %o\t", i, v)
	}
	status += "\n"

	for i, v := range m.PAR {
		status += fmt.Sprintf("PAR[%d]: %o\t", i, v)
	}
	status += "\n"

	return status
}
