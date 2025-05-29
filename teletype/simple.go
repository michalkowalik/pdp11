package teletype

import (
	"fmt"
	"log"
	"os"
	"pdp/interrupts"
	//"pdp/logger"
)

// Simple type  - simplest terminal emulator possible.
type Simple struct {
	KeyboardInput chan uint8

	// TKS : Reader status Register (addr. xxxx560)
	// bits in register:
	// 11: RCVR BUSY, 7: RCVR DONE, 6: RCVR INT ENB, 0: RDR ENB
	TKS uint16

	// TPS : Punch status register (addr. xxxx564)
	// 7: XMT RDY, 6: XMIT INT ENB, 2: MAINT
	TPS uint16

	// ? -> is it a register holding incoming char?
	TKB uint16

	// ??
	TPB uint16

	// ready to receive the next order
	ready bool

	// step delay
	count uint8

	interruptQueue *interrupts.InterruptQueue

	log *log.Logger
}

//var plogger *logger.PLogger

// NewSimple returns the new teletype object
func NewSimple(interruptQueue *interrupts.InterruptQueue, keyboardInput chan uint8, log *log.Logger) *Simple {
	tele := Simple{}
	tele.interruptQueue = interruptQueue

	tele.log = log

	// initialize channels
	tele.KeyboardInput = keyboardInput
	return &tele
}

// GetIncoming returns incoming channel - needed for interface to work
// A placeholder method to keep the interface definition happy
func (t *Simple) GetIncoming() chan Instruction {
	return nil
}

// Run - Start the teletype
// initialize the go routine to read from the incoming channel.
func (t *Simple) Run() error {
	t.ClearTerminal()
	fmt.Printf("Starting teletype terminal\n")
	go t.stdin()
	return nil
}

// Step - single step in terminal operation.
func (t *Simple) Step() {
	if t.ready {
		select {
		case v, ok := <-t.KeyboardInput:
			if ok {
				t.AddChar(v)
			}
		default:
		}
	}

	t.count++
	if t.count%32 != 0 {
		return
	}

	if t.TPS&0x80 == 0 {
		t.writeTerminal(int(t.TPB & 0x7F))
		t.TPS |= 0x80
		if t.TPS&(1<<6) != 0 {
			t.interruptQueue.SendInterrupt(4, interrupts.TTYout)
			t.log.Printf("Sending TTY interrupt %o\n", interrupts.TTYout)
		}
	}
}

func (t *Simple) stdin() {
	for _, v := range []byte("unix\n") {
		t.KeyboardInput <- v
	}

	var b [1]byte
	for {
		n, err := os.Stdin.Read(b[:])
		if n == 1 {
			t.log.Println("Registered keystroke", string(b[:n]))
			t.KeyboardInput <- b[0]
		}
		if err != nil {
			log.Fatal(err)
		}
	}
}

// ClearTerminal - reset terminal
func (t *Simple) ClearTerminal() {
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
		// skip
	default:
		outb[0] = byte(char)
		_, err := os.Stdout.Write(outb[:])
		if err != nil {
			panic("can't write to the output buffer")
		}
	}
}

// getChar - return char from key buffer set registers accordingly
func (t *Simple) getChar() uint16 {
	// fmt.Printf("GET CHAR: TKS:%x, TKB:%x\n", t.TKS, t.TKB)
	if t.TKS&0x80 != 0 {
		t.TKS &= 0xFF7E
		t.ready = true
		return t.TKB
	}
	return 0
}

func (t *Simple) AddChar(char byte) {
	t.log.Println("Adding char", char)
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

	t.TKS |= 0x80
	t.ready = false
	if t.TKS&(1<<6) != 0 {
		t.interruptQueue.SendInterrupt(4, interrupts.TTYin)
	}
}

// WriteTerm : write to the terminal address:
// Warning: Unibus needs to provide a map between the physical 22 bit
// addresses and the 18 bit, DEC defined addresses for the devices.
// TODO: this method can be private!
func (t *Simple) WriteTerm(address uint32, data uint16) error {
	switch address & 0777 {

	// keyboard control & status
	// and why is it never called?
	case 0560:
		if data&(1<<6) != 0 {
			t.TKS |= 1 << 6
		} else {
			t.TKS &= ^(uint16(1 << 6))
		}
	// printer control & status
	case 0564:
		if data&(1<<6) != 0 {
			t.TPS |= 1 << 6
		} else {
			t.TPS &= ^(uint16(1 << 6))
		}
	// output
	// side note:
	// The original implementation introduces 1ms timeouts before setting the register value
	// I'm not sure what it should be good for. anyhow, it looks like it works anyway,
	// so I'm skipping that part.
	case 0566:
		t.TPB = data & 0xFF
		t.TPS &= 0xFF7F
	// any other address -> error
	default:
		panic(fmt.Sprintf("Write to invalid address %o\n", address))
	}
	return nil
}

// ReadTerm - read from terminal memory at address
func (t *Simple) ReadTerm(address uint32) uint16 {
	switch address & 0777 {
	case 0560:
		return t.TKS
	case 0562:
		return t.getChar()
	case 0564:
		return t.TPS
	case 0566:
		return 0
	default:
		panic(fmt.Sprintf("TERM: Read from invalid address: %o\n", address))
	}
}
