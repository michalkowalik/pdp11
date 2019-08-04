package unibus

import (
	"errors"
	"fmt"
	"io/ioutil"
	"pdp/interrupts"
)

const (
	// RKDEBUG flag to output extra info
	RKDEBUG = false

	rk5ImageLength = 2077696
	// unibus Addresses:
	rkdsAddress = 0777400
	rkerAddress = 0777402
	rkcsAddress = 0777404
	rkwcAddress = 0777406
	rkbaAddress = 0777410
	rkdaAddress = 0777412

	// RK11 error codes:
	rkOvr = (1 << 14)
	rkNxd = (1 << 7)
	rkNxc = (1 << 6)
	rkNxs = (1 << 5)
)

// RK11 disk controller
type RK11 struct {
	// Registers -> check description in attached markdown
	RKDS uint16
	RKER uint16
	RKCS uint16

	// we do want to have RKWC as signed integer, as it is for whatever reason
	// counting up to zero starting with negative value. not sure what were they
	// smoking at DEC back then.
	RKWC int
	RKBA uint16
	DKDA uint16

	// disk units
	unit [8]*RK05

	//disk geometry:
	// TODO: make sure "int" is the most fitting type. Perhaps uint16 is good enough.
	drive, sector, surface, cylinder int

	running bool

	unibus *Unibus
}

// RK05 disk cartridge
type RK05 struct {
	rdisk  []byte
	locked bool
}

// Instruction - to provide unibus exchange channel
// Read == false -> write operation
// TODO: it slowly looks like teletype and rk could implement the same interface.
type Instruction struct {
	Address uint32
	Data    uint16
	Read    bool
}

// NewRK  returns new RK11 object
func NewRK(u *Unibus) *RK11 {
	r := RK11{}
	r.unibus = u
	return &r
}

// Attach reads disk image file and loads it to memory
func (r *RK11) Attach(drive int, path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	unit := &RK05{
		rdisk: buf,
	}

	if drive >= len(r.unit) {
		return errors.New("tried to mount disk to unit > 7")
	}

	r.unit[drive] = unit

	r.unibus.controlConsole.WriteConsole("disk mounted")

	return nil
}

// rkReady - set Drive Ready bit in RKDS and Control Ready bit in RKCS registers to 1
func (r *RK11) rkReady() {
	r.RKDS |= 1 << 6
	r.RKCS |= 1 << 7
}

// rkNotReady - set Drive Ready bit in RKDS and Control Ready bit in RKCS registers to 0
func (r *RK11) rkNotReady() {
	r.RKDS = r.RKDS &^ (1 << 6)
	r.RKCS = r.RKCS &^ (1 << 7)
}

// read and return drive register value
func (r *RK11) read(address uint32) uint16 {
	if RKDEBUG {
		fmt.Printf("RK: Reading from address %o\n", address)
	}
	switch address {
	case rkdsAddress:
		return r.RKDS
	case rkerAddress:
		return r.RKER
	case rkcsAddress:
		return r.RKCS
	case rkwcAddress:
		return uint16(r.RKWC)
	case rkbaAddress:
		return r.RKBA
	case rkdaAddress:
		return uint16(r.sector | (r.surface << 4) | (r.cylinder << 5) | (r.drive << 13))
	default:
		panic("invalid RK11 read")
	}
}

func (r *RK11) write(address uint32, value uint16) {
	switch address {
	case rkdsAddress:
		break
	case rkerAddress:
		break
	case rkcsAddress:
		// set bus address:
		r.RKBA = r.RKBA | ((value & 060) << 12)

		var bits uint16 = 017517

		// set only the writeable bits:
		value &= bits
		r.RKCS &= ^bits

		// don't set the GO bit
		r.RKCS |= value & ^uint16(1)
		if value&1 == 1 {
			r.rkgo()
		}
	case rkwcAddress:
		r.RKWC = int(int(^value+1) * -1)
	case rkbaAddress:
		r.RKBA = value
	case rkdaAddress:
		r.drive = int(value >> 13)
		r.cylinder = int(value>>5) & 0377
		r.surface = int(value>>4) & 1
		r.sector = int(value & 15)
	default:
		panic("RK5: Invalid write")
	}

}

// Respond to GO bit set in RKCS - start operations
func (r *RK11) rkgo() {
	if RKDEBUG {
		fmt.Printf("RK: It's a go, all engines running!\n")
		fmt.Printf("RKWC: %o\n", r.RKWC)
	}
	switch (r.RKCS & 017) >> 1 {
	case 0: // Control reset
		r.Reset()
	case 1, 2: // R/W
		r.running = true
		r.rkNotReady()
	case 7: // write lock
		r.running = true
		r.rkNotReady()
	default:
		panic(fmt.Sprintf("unimplemented RK5 operation %#o", ((r.RKCS & 017) >> 1)))
	}
}

// Reset sets the drive to it's default values.
// check bits meaning in attached documentation
func (r *RK11) Reset() {
	r.RKDS = (1 << 11) | (1 << 7) | (1 << 6)
	r.RKER = 0
	r.RKCS = 1 << 7
	r.RKWC = 0
	r.RKBA = 0
}

// rkerror is being called in response to specific RK11 error
func (r *RK11) rkError(code uint16) {
	var msg string

	r.rkReady()
	r.RKER |= code
	r.RKCS |= (1 << 14) | (1 << 15)

	switch code {
	case rkOvr:
		msg = "operation overflowed the disk"
		break
	case rkNxd:
		msg = "invalid disk accessed"
		break
	case rkNxc:
		msg = "invalid cylinder accessed"
		break
	case rkNxs:
		msg = "invalid sector accessed"
		break
	}
	panic(msg)
}

// Step - single operation step
func (r *RK11) Step() {
	if !r.running {
		return
	}

	if r.unit[r.drive] == nil {
		r.rkError(rkNxd)
	}

	var isWrite bool
	unit := r.unit[r.drive]

	// check the "function" fields in RKCS register
	switch (r.RKCS & 017) >> 1 {
	case 01:
		isWrite = true
	case 02:
		isWrite = false
	case 07:
		unit.locked = true
		r.running = false
		r.rkReady()
		return
	default:
		panic(fmt.Sprintf("unimplemented RK05 operation: %#o", ((r.RKCS & 017) >> 1)))
	}

	// set the head location:
	if r.cylinder > 0312 {
		r.rkError(rkNxc)
	}
	if r.sector > 013 {
		r.rkError(rkNxc)
	}
	pos := (r.cylinder*24 + r.surface*12 + r.sector) * 512
	if pos >= len(unit.rdisk) {
		panic(fmt.Sprintf("pos outside rkdisk length, pos: %v, len %v", pos, len(r.unit[r.drive].rdisk)))
	}

	// reaad complete sector:
	for i := 0; i < 256 && r.RKWC != 0; i++ {
		if isWrite {
			val := r.unibus.Mmu.ReadMemoryWord(r.RKBA)
			unit.rdisk[pos] = byte(val & 0xFF)
			unit.rdisk[pos+1] = byte((val >> 8) & 0xFF)
		} else {
			if RKDEBUG {
				fmt.Printf("RK read: RKBA: %o, RKWC: %d, Position: %o\n", r.RKBA, r.RKWC, pos)
			}
			// TODO: monitor if it's fine. this implementation does not take care of
			// bits 4 and 5 of rkcs, which should be used on systems with extended memory
			r.unibus.Mmu.WriteMemoryWord(
				r.RKBA,
				uint16(unit.rdisk[pos])|uint16(unit.rdisk[pos+1])<<8)
		}
		r.RKBA += 2
		pos += 2
		r.RKWC++
	}
	r.sector++
	if r.sector > 13 {
		r.sector = 0
		r.surface++
		if r.surface > 1 {
			r.surface = 0
			r.cylinder++
			if r.cylinder > 0312 {
				r.rkError(rkOvr)
			}
		}
	}

	// RKWC == 0 -> transfer is completed. if bit 6 set in RKCS, interrupt should be sent.
	if r.RKWC == 0 {
		fmt.Printf("RKWC: 0, transfer complete, RKCS: %o\n", r.RKCS)
		r.running = false
		r.rkReady()
		if r.RKCS&(1<<6) != 0 {
			r.unibus.SendInterrupt(5, interrupts.INTRK)
		}
	}
}
