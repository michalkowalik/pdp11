package unibus

// Interrupt type - used to sygnalize incoming interrupt
// perhaps to be added:
// - delay
// - callback
// - callarg
type Interrupt struct {
	Priority  uint16
	vector    uint16
	cleanFlag bool
}

// Unibus definition
type Unibus struct {

	// Unibus map registers
	UnibusMap [32]int16

	// Channel for interrupt communication
	Interrupts chan Interrupt

	// Unibus memory map. rough and uncomplete for now.
	// the variadic parameters are required due to some
	// of the IOPage function requireding 4 parameters
	// TODO: 4'th parameter being...
	memoryMap map[uint32](func(bool, ...uint32) error)
}

// New initializes and returns the Unibus variable
func New() *Unibus {
	unibus := Unibus{}
	unibus.Interrupts = make(chan Interrupt)

	// initialize memory map:
	unibus.memoryMap = make(map[uint32](func(bool, ...uint32) error))
	unibus.memoryMap[017772000] = unibus.accessVT100

	return &unibus
}

// map 18 bit unibus address to 22 bit physical via the unibus map (if active)
//  TODO: implementation missing
func (u *Unibus) mapUnibusAddress(unibusAddress uint32) uint32 {
	return 0
}

// access IO Page - write or read.
// TODO: implementation.
func (u *Unibus) accessIOPage(physicalAddres uint32, data uint16, byteFlag bool) error {
	if val, ok := u.memoryMap[physicalAddres]; ok {
		val(byteFlag, physicalAddres, uint32(data))
	}
	return nil
}

// SendInterrupt sends a new interrupts to the receiver
// TODO: implementation!
func (u *Unibus) SendInterrupt(priority uint16, vector uint16) {
	i := Interrupt{
		Priority: priority,
		vector:   vector}

	// send interrupt:
	go func() { u.Interrupts <- i }()
}
