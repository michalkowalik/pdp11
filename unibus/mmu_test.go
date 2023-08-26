package unibus

import (
	"pdp/console"
	"pdp/psw"
	"testing"
)

func TestMMU_DecodeAddress(t *testing.T) {
	p := psw.PSW(0)
	var cons console.Console = console.NewSimple()
	u := New(&p, nil, &cons, false)

	// enable mmu:
	u.Mmu.SetSR0(u.Mmu.GetSR0() | 1)

	var virtualAddress uint16 = 0157746
	var physicalAddress Uint18 = 0565746
	var relocationConstant uint16 = 05460

	// set the PAF in PAR:
	u.Mmu.Write16(0772354, relocationConstant)

	// set read, ed = 0 and page length  in PDR:
	u.Mmu.Write16(0772314, 077606)

	t.Run("Decode 16 bit virtual address to 18 bit physical", func(t *testing.T) {
		decodedAddress := u.Mmu.Decode(virtualAddress, false, false)
		if decodedAddress != physicalAddress {
			t.Errorf("Expected decoded address to equal %06o, got %06o", physicalAddress, decodedAddress)
		}
	})
}
