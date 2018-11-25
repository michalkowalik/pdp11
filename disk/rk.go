package disk

import (
	"io/ioutil"
)

const rk5ImageLength = 2077696

// RK11 disk controller
type RK11 struct {
	// Registers -> check description in attached markdown
	RKDS uint16
	RKER uint16
	RKCS uint16
	RKWC uint16
	RKBA uint16

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
// btw, it slowly looks like teletype and rk could implement the same interface.
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
	r.unit[drive] = unit
	return nil
}

// Run - Start the RK11
// initialize the go routine to read from incoming channel
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

func (r *RK11) rkReady() {
	// nothing to see here yet
}

func (r *RK11) rkNotReady() {
	// nothing to see here yet
}

func (r *RK11) read(address uint32) uint16 {
	return 0
}

func (r *RK11) write(address uint32, value uint16) {
	// end up in panic?
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
