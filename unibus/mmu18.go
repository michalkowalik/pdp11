package unibus

import (
	"fmt"
	"pdp/interrupts"
)

// I think it acutally is being called KT11
const DEBUG_MMU = false

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

func (m *MMU18) Read16(addr uint18) uint16 {
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
		Vector: interrupts.INTBus,
		Msg:    fmt.Sprintf("Attempt to read from invalid address %06o", addr)})

}

func (m *MMU18) ReadMemoryWord(a uint16) uint16 {
	// TODO:Finish
	return 0
}

func (m *MMU18) ReadMemoryByte(a uint16) byte {
	//TODO: finish
	return 0
}

func (m *MMU18) Write16(addr uint18, data uint16) {
	i := ((addr & 017) >> 1)
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
		Vector: interrupts.INTBus,
		Msg:    fmt.Sprintf("Attempt to read from invalid address %06o", addr)})
}

func (m *MMU18) WriteMemoryWord(addr, data uint16) {
	// TODO:Finish
}

func (m *MMU18) WriteMemoryByte(addr uint16, data byte) {
	//TODO: finish
}

func (m *MMU18) MmuEnabled() bool {
	return m.SR0&1 == 1
}

func (m *MMU18) Decode(a uint16, w, user bool) (addr uint18) {
	if !m.MmuEnabled() {
		aa := uint18(a)
		if aa >= 0170000 {
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
			Vector: interrupts.INTFault,
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
			Vector: interrupts.INTFault,
			Msg:    fmt.Sprintf("read from no-access page %06o", a),
		})
	}
	block := (a >> 6) & 0177
	disp := uint18(a & 077)
	if p.ed() && block < p.len() || !p.ed() && block > p.len() {
		//if(p.ed ? (block < p.len) : (block > p.len)) {
		m.SR0 = (1 << 14) | 1
		m.SR0 |= (a >> 12) & ^uint16(1)
		if user {
			m.SR0 |= (1 << 5) | (1 << 6)
		}
		m.SR2 = m.unibus.PdpCPU.Registers[7]
		panic(interrupts.Trap{
			Vector: interrupts.INTFault,
			Msg: fmt.Sprintf("page length exceeded, address %06o (block %03o) is beyond %03o",
				a, block, p.len())})
	}
	if w {
		p.pdr |= 1 << 6
	}
	aa := ((uint18(block) + uint18(p.addr())) << 6) + disp
	if DEBUG_MMU {
		fmt.Printf("decode: slow %06o -> %06o\n", a, aa)
	}
	return aa

}

func (m *MMU18) DumpMemory() {
	// TODO: finish
}

func (m *MMU18) GetSR0() uint16 {
	return m.SR0
}

func (m *MMU18) GetSR2() uint16 {
	return m.SR2
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
