package mmu

import "pdp/unibus"

// I think it acutally is being called KT11

type page struct {
	par, pdr uint16
}

type mmu18 struct {
	SR0, SR2 uint16
	pages    [16]page
	unibus   *unibus.Unibus
}

func (p *page) read() bool   { return p.pdr&2 == 2 }
func (p *page) write() bool  { return p.pdr&6 == 6 }
func (p *page) ed() bool     { return p.pdr&8 == 8 }
func (p *page) addr() uint16 { return p.par & 07777 }
func (p *page) len() uint16  { return (p.pdr >> 8) & 0x7f }

func (m *mmu18) Read16(addr uint18) uint16 {
	// TODO: finish
	return 0
}

func (m *mmu18) Write16(addr uint18, value uint16) {
	// nothing to see here yet
}

func (m *mmu18) MmuEnabled() bool {
	// TODO: finish
	return true
}

func (m *mmu18) Decode(a uint16, w, user bool) (addr uint18) {
	return 0
}
