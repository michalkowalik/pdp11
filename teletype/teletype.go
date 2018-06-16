package teletype

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

	// incoming and outgoing stream
	// incoming for the events coming from the unibus:
	// -> print char (and in this case it's probably everthing,
	//    as there's not delete, no refresh and new line is basically
	//    just another char)
	// outgoing channel for communicating with unibus:
	// trigger interrupt when sending a char

	// Incoming channel
	Incoming chan uint16

	// Outgoing channel
	Outgoing chan uint16
}

// New returns new teletype object
func New(termView *gocui.View) *Teletype {
	tele := Teletype{}
	tele.termView = termView

	// initialize channels
	tele.Incoming = make(chan uint16, 8)
	tele.Outgoing = make(chan uint16, 8)

	return &tele
}

// Run : Start the teletype
func (t *Teletype) Run() error {
	return nil
}

// ReadTerm : read from terminal memory at address address
func (t *Teletype) ReadTerm(address uint32) (uint16, error) {
	return 0, nil
}

// WriteTerm : write to the terminal address:
func (t *Teletype) WriteTerm(address uint32, data uint16) error {
	return nil
}
