package unibus

import (
	"fmt"
	"pdp/interrupts"
)

const DEBUG_MMU = false

// RegisterAddressVirtual - begin of register address in 16 bit space
const RegisterAddressVirtual = 0177700

type MMU18 struct {
	SR0, SR2 uint16
	pages    [16]page
	unibus   *Unibus
}

func (p *page) read() bool   { return p.pdr&2 == 2 }
func (p *page) write() bool  { return p.pdr&6 == 6 }
func (p *page) ed() bool     { return p.pdr&8 == 8 }
func (p *page) addr() uint16 { return p.par & 07777 }
func (p *page) len() uint16  { return (p.pdr >> 8) & 0x7f }

func (m *MMU18) Read16(addr Uint18) uint16 {
	i := (addr & 017) >> 1

	// kernel space:
	if (addr >= 0772300) && (addr < 0772320) {
		return m.pages[i].pdr
	}
	if (addr >= 0772340) && (addr < 0772360) {
		return m.pages[i].par
	}

	// user space:
	if (addr >= 0777600) && (addr < 0777620) {
		return m.pages[i+8].pdr
	}
	if (addr >= 0777640) && (addr < 0777660) {
		return m.pages[i+8].par
	}
	panic(interrupts.Trap{
		Vector: interrupts.IntBUS,
		Msg:    fmt.Sprintf("Attempt to read from invalid address %06o", addr)})

}

func (m *MMU18) ReadMemoryWord(a uint16) uint16 {
	if a&0177770 == RegisterAddressVirtual {
		return m.unibus.PdpCPU.Registers[a&7]
	}

	pAddr := m.Decode(a, false, m.unibus.PdpCPU.IsUserMode())
	return m.unibus.ReadIO(pAddr)
}

func (m *MMU18) ReadMemoryByte(a uint16) byte {
	if a&0177770 == RegisterAddressVirtual {
		return byte(m.unibus.PdpCPU.Registers[a&7] & 0xff)
	}

	pAddr := m.Decode(a, false, m.unibus.PdpCPU.IsUserMode())
	return byte(m.unibus.ReadIOByte(pAddr))
}

func (m *MMU18) Write16(addr Uint18, data uint16) {
	i := (addr & 017) >> 1
	if (addr >= 0772300) && (addr < 0772320) {
		m.pages[i].pdr = data
		return
	}
	if (addr >= 0772340) && (addr < 0772360) {
		m.pages[i].par = data
		return
	}
	if (addr >= 0777600) && (addr < 0777620) {
		m.pages[i+8].pdr = data
		return
	}
	if (addr >= 0777640) && (addr < 0777660) {
		m.pages[i+8].par = data
		return
	}
	panic(interrupts.Trap{
		Vector: interrupts.IntBUS,
		Msg:    fmt.Sprintf("Attempt to write to an invalid address %06o", addr)})
}

func (m *MMU18) WriteMemoryWord(addr, data uint16) {
	if addr&0177770 == RegisterAddressVirtual {
		m.unibus.PdpCPU.Registers[addr&7] = data
		return
	}

	pAddr := m.Decode(addr, true, m.unibus.PdpCPU.IsUserMode())
	m.unibus.WriteIO(pAddr, data)
}

func (m *MMU18) WriteMemoryByte(addr uint16, data byte) {
	// modify register directly:
	if (addr & 0177770) == 0177700 {
		m.unibus.PdpCPU.Registers[addr&7] = uint16(data)
		return
	}

	pAddr := m.Decode(addr, true, m.unibus.PdpCPU.IsUserMode())
	m.unibus.WriteIOByte(pAddr, uint16(data))
}

func (m *MMU18) MmuEnabled() bool {
	return m.SR0&1 == 1
}

// Decode 16 bit virtual address to 18 bit physical address
func (m *MMU18) Decode(a uint16, w, user bool) (addr Uint18) {
	if !m.MmuEnabled() {
		aa := Uint18(a)
		if aa >= 0170000 { // unibus memory address space begin
			aa += 0600000
		}
		if DEBUG_MMU {
			fmt.Printf("decode: fast %06o -> %06o\n", a, aa)
		}
		return aa
	}
	offset := a >> 13
	if user {
		offset += 8
	}
	p := m.pages[offset]
	if w && !p.write() {
		m.SR0 = (1 << 13) | 1
		m.SR0 |= a >> 12 & ^uint16(1)
		if user {
			m.SR0 |= (1 << 5) | (1 << 6)
		}
		m.SR2 = m.unibus.PdpCPU.Registers[7]
		panic(interrupts.Trap{
			Vector: interrupts.IntFAULT,
			Msg:    fmt.Sprintf("Abort: write on read-only page %o\n", a)})
	}
	if !p.read() {
		m.SR0 = (1 << 15) | 1
		m.SR0 |= (a >> 12) & ^uint16(1)
		if user {
			m.SR0 |= (1 << 5) | (1 << 6)
		}
		m.SR2 = m.unibus.PdpCPU.Registers[7]
		panic(interrupts.Trap{
			Vector: interrupts.IntFAULT,
			Msg:    fmt.Sprintf("read from no-access page %06o", a),
		})
	}
	block := (a >> 6) & 0177
	disp := Uint18(a & 077)
	if p.ed() && block < p.len() || !p.ed() && block > p.len() {
		m.SR0 = (1 << 14) | 1
		m.SR0 |= (a >> 12) & ^uint16(1)
		if user {
			m.SR0 |= (1 << 5) | (1 << 6)
		}
		m.SR2 = m.unibus.PdpCPU.Registers[7]
		panic(interrupts.Trap{
			Vector: interrupts.IntFAULT,
			Msg: fmt.Sprintf("page length exceeded, address %06o (block %03o) is beyond %03o",
				a, block, p.len())})
	}
	if w {
		p.pdr |= 1 << 6
	}
	aa := ((Uint18(block) + Uint18(p.addr())) << 6) + disp
	if DEBUG_MMU {
		fmt.Printf("decode: slow %06o -> %06o\n", a, aa)
	}
	return aa

}

func (m *MMU18) GetSR0() uint16 {
	return m.SR0
}

func (m *MMU18) GetSR2() uint16 {
	return m.SR2
}

func (m *MMU18) SetSR0(v uint16) {
	m.SR0 = v
}

func (m *MMU18) SetSR2(v uint16) {
	m.SR2 = v
}

func (m *MMU18) SetPage(i int, p page) {
	m.pages[i] = p
}

func NewMMU18(unibus *Unibus) *MMU18 {
	mmu := MMU18{}
	mmu.unibus = unibus
	return &mmu
}
