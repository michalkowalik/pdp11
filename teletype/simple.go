package teletype

import (
	"fmt"
	"log"
	"os"
	"pdp/interrupts"
)

// Simple type  - simplest terminal emulator possible.
type Simple struct {
	keyboardInput chan uint8

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

	interrupts chan interrupts.Interrupt

	// ready to receive next order
	ready bool
}

// NewSimple returns new teletype object
func NewSimple(interrupts chan interrupts.Interrupt) *Simple {
	tele := Simple{}
	tele.interrupts = interrupts

	// initialize channels
	tele.keyboardInput = make(chan uint8)
	return &tele
}

// GetIncoming returns incoming channel - needed for interface to work
// dummy method to keep the interface definition happy
func (t *Simple) GetIncoming() chan Instruction {
	return nil
}

// Run : Start the teletype
// initialize the go routine to read from the incoming channel.
func (t *Simple) Run() error {
	t.clearTerminal()
	fmt.Printf("starting teletype terminal\n")
	go t.stdin()
	return nil
}

// Step - single step in terminal operation.
func (t *Simple) Step() {
	if t.ready {
		select {
		case v, ok := <-t.keyboardInput:
			if ok {
				t.addChar(byte(v))
			}
		default:
		}
	}
	// count == > do I need it at all?
	if t.TPS&0x80 == 0 {
		t.writeTerminal(int(t.TPB & 0x7F))
		t.TPS |= 0x80
		if t.TPS&(1<<6) != 0 {
			t.interrupts <- interrupts.Interrupt{
				Priority: 4,
				Vector:   interrupts.TTYout}
		}
	}
}

func (t *Simple) stdin() {
	var b [1]byte
	for {
		n, err := os.Stdin.Read(b[:])
		if n == 1 {
			t.keyboardInput <- b[0]
		}
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (t *Simple) clearTerminal() {
	t.TKS = 0
	t.TPS = 1 << 7
	t.TKB = 0
	t.TPB = 0
	t.ready = true
}

func (t *Simple) writeTerminal(char int) {
	var outb [1]byte

	switch char {
	case 13:
		//skip
	default:
		outb[0] = byte(char)
		os.Stdout.Write(outb[:])
	}
}

//getChar - return char from keybuffer set registers accordingly
func (t *Simple) getChar() uint16 {
	if t.TKS&0x80 != 0 {
		t.TKS &= 0xFF7E
		t.ready = true
		return t.TKB
	}
	return 0
}

func (t *Simple) addChar(char byte) {
	fmt.Printf("adding char %v\n", char)
	switch char {
	case 42:
		t.TKB = 4
	case 19:
		t.TKB = 034
	case '\n':
		t.TKB = '\r'
	default:
		t.TKB = uint16(char)
	}

	// fmt.Printf("DEBUG: TKB: %x\n", t.TKB)

	t.TKS |= 0x80
	// fmt.Printf("DEBUG: TKS: %x\n", t.TKS)
	t.ready = false
	if t.TKS&(1<<6) != 0 {
		t.interrupts <- interrupts.Interrupt{
			Priority: 4,
			Vector:   interrupts.TTYin}
	}
}

// WriteTerm : write to the terminal address:
// Warning: Unibus needs to provide a map between the physical 22 bit
// addresses and the 18 bit, DEC defined addresses for the devices.
// TODO: this method can be private!
func (t *Simple) WriteTerm(address uint32, data uint16) error {
	// fmt.Printf("DEBUG: Console Write to addr %o\n", address)

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
		//fmt.Printf("DEBUG TELETYPE OUT: %v, TPS: %v\n", data, t.TPS)

		t.TPB = data & 0xFF
		t.TPS &= 0xFF7F
	// any other address -> error
	default:
		panic(fmt.Sprintf("Write to invalid address %o\n", address))
	}
	return nil
}

// ReadTerm : read from terminal memory at address address
func (t *Simple) ReadTerm(address uint32) (uint16, error) {
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
