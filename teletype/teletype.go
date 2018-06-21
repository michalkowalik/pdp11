package teletype

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

// Instruction - Incomming instruction type
// if Read is set to false -> write instruction
type Instruction struct {
	Address uint32
	Data    uint16
	Read    bool
}

// Teletype type  - simplest terminal emulator possible.
type Teletype struct {
	// use gocui view, not a raw terminal
	termView *gocui.View

	// not really sure if it is a good choice for the input type
	keybuffer uint16

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

	fmt.Fprintf(termView, "-Teletype Initialized-\n")
	return &tele
}

// Run : Start the teletype
// initialize the go routine to read from the incoming channel.
func (t *Teletype) Run() error {
	t.TKS = 0
	t.TPS = 1 << 7
	go func() error {
		for {
			select {
			case instruction := <-t.Incoming:
				if instruction.Read {
					data, err := t.ReadTerm(instruction.Address)
					if err != nil {
						return err
					}
					t.Outgoing <- data
				} else {
					err := t.WriteTerm(instruction.Address, instruction.Data)
					if err != nil {
						return err
					}
				}
			default:
			}
		}
	}()
	return nil
}

//getChar - return char from keybuffer set registers accordingly
func (t *Teletype) getChar() uint16 {
	if t.TKS&0x80 != 0 {
		t.TKS &= 0xFF7E
		return t.keybuffer
	}
	return 0
}

// WriteTerm : write to the terminal address:
// Warning: Unibus needs to provide a map between the physical 22 bit
// addresses and the 18 bit, DEC defined addresses for the devices.
func (t *Teletype) WriteTerm(address uint32, data uint16) error {
	switch address & 0777 {

	// keyboard control & status
	case 0560:
		if data&(1<<6) != 0 {
			t.TKS |= 1 << 6
		} else {
			t.TKS &= ^(uint16(1 << 6))
		}
		break

	// printer control & status
	case 0564:
		if data&(1<<6) != 0 {
			t.TPS |= 1 << 6
		} else {
			t.TPS &= ^(uint16(1 << 6))
		}
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
		break

		// any other address -> error
	default:
		return fmt.Errorf("Write to invalid address %o", address)
	}
	return nil
}

// ReadTerm : read from terminal memory at address address
func (t *Teletype) ReadTerm(address uint32) (uint16, error) {
	switch address & 0777 {
	case 0560:
		return t.TKS, nil
	case 0562:
		return t.getChar(), nil
	case 0564:
		return t.TPS, nil
	case 0566:
		return 0, nil
	default:
		return 0, fmt.Errorf("Read from invalid address: %o", address)
	}
}
