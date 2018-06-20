package teletype

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

// Instruction - Incomming instruction type
type Instruction struct {
	Address uint16
	Data    uint16
}

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

	// TPS : Punch status register (addr. xxxx564)
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
	Incoming chan Instruction

	// Outgoing channel
	Outgoing chan uint16
}

// New returns new teletype object
func New(termView *gocui.View) *Teletype {
	tele := Teletype{}
	tele.termView = termView

	// initialize channels
	tele.Incoming = make(chan Instruction, 8)

	// outgoing channel is boud to trigger the interrupt -
	// the type needs to be changed probably as well.
	tele.Outgoing = make(chan uint16, 8)

	return &tele
}

// Run : Start the teletype
// initialize the go routine to read from the incoming channel.
func (t *Teletype) Run() error {
	return nil
}

// ReadTerm : read from terminal memory at address address
func (t *Teletype) ReadTerm(address uint32) (uint16, error) {
	return 0, nil
}

// WriteTerm : write to the terminal address:
// Warning: Unibus needs to provide a map between the physical 22 bit
// addresses and the 18 bit, DEC defined addresses for the devices.
func (t *Teletype) WriteTerm(address uint32, data uint16) error {
	switch address & 0777 {

	// keyboard control & status
	case 0560:
		break

	// printer controal & status
	case 0564:
		break

	// output
	case 0566:
		data = data & 0xFF
		if t.TPS&0x80 == 0 {
			break
		}
		if data == 13 {
			break
		}
		fmt.Fprint(t.termView, string(data&0x7F))
		t.TPS &= 0xFF7F
		// need timeouts here!

		// any other address -> error
	default:
		return fmt.Errorf("Write to invalid address %v", address)
	}
	return nil
}
