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
}

// New initializes and returns the Unibus variable
func New() *Unibus {
	unibus := Unibus{}
	unibus.Interrupts = make(chan Interrupt)
	return &unibus
}

func (u *Unibus) accessIOPage(physicalAddres uint32, data uint16, byteFlag bool) error {
	return nil
}

// SendInterrupt sends a new interrupts to the receiver
func (u *Unibus) SendInterrupt(priority uint16, vector uint16) {
	i := Interrupt{
		Priority: priority,
		vector:   vector}

	// send interrupt:
	go func() { u.Interrupts <- i }()
}
