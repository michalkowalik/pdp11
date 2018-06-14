package unibus

import "github.com/jroimartin/gocui"

// Teletype type  - simplest terminal emulator possible.
type Teletype struct {
	// use gocui view, not a raw terminal
	termView *gocui.View

	// not really sure if it is a good choice for the input type
	// check with the gocui
	keybuffer rune

	// TKS : Reader status Register (addr. xxxx560)
	// bits in register:
	// 11: RCVR BUSY, 7: RCRV DONE, 6: RCVR INT ENB, 0: RDR ENB
	TKS uint16

	// TPS : Punch status register
	// 7: XMT RDY, 6: XMIT INT ENB, 2: MAINT
	TPS uint16
}

// New returns new teletype object
func New(termView *gocui.View) *Teletype {
	tele := Teletype{}
	tele.termView = termView
	return &tele
}
