package disk

import (
	"errors"
	"fmt"
	"io/ioutil"
)

const (
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
	RKWC uint16
	RKBA uint16
	DKDA uint16

	// disk units
	unit [8]*RK05

	//disk geometry:
	// TODO: make sure "int" is the most fitting type. Perhaps uint16 is good enough.
	drive, sector, surface, cylinder int

	running bool

	// we also somehow need the unibus here...
	Instructions chan Instruction
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

// New  returns new RK11 object
func New() *RK11 {
	r := RK11{}
	r.Instructions = make(chan Instruction)
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
	return nil
}

// Run - Start the RK11
// initialize the go routine to read from incoming channel
// TODO: goroutine should be calling the Step function!
func (r *RK11) Run() {
	go func() {
		instruction := <-r.Instructions
		if instruction.Read {
			//data := r.read(instruction.Address)
			r.read(instruction.Address)
			// send back to Unibus!
		} else {
			r.write(instruction.Address, instruction.Data)
		}
	}()
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
	switch address {
	case rkdsAddress:
		return r.RKDS
	case rkerAddress:
		return r.RKER
	case rkcsAddress:
		return r.RKCS
	case rkwcAddress:
		return r.RKWC
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
		r.RKWC = value
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
	switch (r.RKCS & 017) >> 1 {
	case 0: // Control reset
		r.reset()
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

// reset sets the drive to it's default values.
// check bits meaning in attached documentation
func (r *RK11) reset() {
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

	// check the "function" fields in RKCS register
	switch (r.RKCS & 017) >> 1 {
	case 01:
		isWrite = true
	case 02:
		isWrite = false
	case 07:
		r.unit[r.drive].locked = true
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
	if pos >= len(r.unit[r.drive].rdisk) {
		panic(fmt.Sprintf("pos outside rkdisk length, pos: %v, len %v", pos, len(r.unit[r.drive].rdisk)))
	}

	// reaad complete sector:
	for i := 0; i < 256 && r.RKWC != 0; i++ {
		if isWrite {
			// write to the disk
			fmt.Printf("write to disk \n")
		} else {

		}
	}

}
