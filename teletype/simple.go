package teletype

import (
	"fmt"
	"os"
	"pdp/interrupts"
	"time"
)

// Simple type  - simplest terminal emulator possible.
type Simple struct {
	// not really sure if it is a good choice for the input type
	keybuffer uint16

	// TKS : Reader status Register (addr. xxxx560)
	// bits in register:
	// 11: RCVR BUSY, 7: RCRV DONE, 6: RCVR INT ENB, 0: RDR ENB
	TKS uint16

	// TPS : Punch status register (addr. xxxx564)
	// 7: XMT RDY, 6: XMIT INT ENB, 2: MAINT
	TPS uint16

	// ? -> is it a register holding incoming char?
	TKB uint16

	// ??
	TPB uint16

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

	// keystroke channel
	keystrokes chan rune

	// terminal out channel -> required, as due to way gocui refreshes the
	// view, it needs to happen in the separate goroutine
	consoleOut chan string

	interrupts chan interrupts.Interrupt

	// trying to synchronize the output...
	done chan bool
}

// NewSimple returns new teletype object
func NewSimple(
	interrupts chan interrupts.Interrupt) *Simple {
	tele := Simple{}
	tele.interrupts = interrupts

	// initialize channels
	tele.Incoming = make(chan Instruction)

	// outgoing channel is bound to trigger the interrupt -
	// the type needs to be changed probably as well.
	tele.Outgoing = make(chan uint16)
	tele.keystrokes = make(chan rune)
	tele.consoleOut = make(chan string)
	tele.done = make(chan bool)
	tele.initOutput()
	return &tele
}

// GetIncoming returns incoming channel - needed for interface to work
func (t *Simple) GetIncoming() chan Instruction {
	return t.Incoming
}

// add char ->

// initOutput starts a goroutine reading from the consoleOut channel
// and calling the gocui.gui.Update to modify the view.
// t.done channel is used to force synchronization.
// 1ms sleep is required to give the gocui enough time to print the character.
// consult gocui documentation for further details.
func (t *Simple) initOutput() {
	go func() {
		for {
			s := <-t.consoleOut
			os.Stdout.Write([]byte(s))
			time.Sleep(1 * time.Millisecond)
			t.done <- true
		}
	}()
}

// Run : Start the teletype
// initialize the go routine to read from the incoming channel.
func (t *Simple) Run() error {
	t.TKS = 0
	t.TPS = 1 << 7
	fmt.Printf("starting console\n")
	go func() error {
		for {
			instruction := <-t.Incoming
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
		}
	}()
	t.consoleOut <- "-Teletype Initialized-\n"
	<-t.done
	return nil
}

//getChar - return char from keybuffer set registers accordingly
func (t *Simple) getChar() uint16 {
	if t.TKS&0x80 != 0 {
		t.TKS &= 0xFF7E
		return t.keybuffer
	}
	return 0
}

// WriteTerm : write to the terminal address:
// Warning: Unibus needs to provide a map between the physical 22 bit
// addresses and the 18 bit, DEC defined addresses for the devices.
// TODO: this method can be private!
func (t *Simple) WriteTerm(address uint32, data uint16) error {
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
	// side note:
	// The original implementation introduces 1ms timeouts before setting the register value
	// I'm not sure what should it be good for. anyhow, it looks like it works anyway,
	// so I'm skipping that part.
	case 0566:
		fmt.Printf("DEBUG TELETYPE OUT: %v, TPS: %v\n", data, t.TPS)
		data = data & 0xFF
		if t.TPS&0x80 == 0 {
			fmt.Printf("breaking! data: %d, tps: %x\n", data, t.TPS)
			break
		}
		if data == 13 {
			break
		}
		t.consoleOut <- string(data & 0x7F)
		<-t.done

		t.TPS &= 0xFF7F
		t.TPS = t.TPS | 0x80
		if t.TPS&(1<<6) != 0 {
			// send interrupt
			t.interrupts <- interrupts.Interrupt{
				Priority: 4,
				Vector:   interrupts.TTYout}
		}
		break

	// any other address -> error
	default:
		return fmt.Errorf("Write to invalid address %o", address)
	}
	return nil
}

// ReadTerm : read from terminal memory at address address
func (t *Simple) ReadTerm(address uint32) (uint16, error) {
	fmt.Printf("Attempting to read from teletype from address %v \n", address)

	/*
		TODO: - why does it want ro read from terminal, if there was no attempt to write to it?
			  - why does it want to read from character buffer??
			  - check the debug output what has happened there.
			  - have I got the addresses wrong?
	*/

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
