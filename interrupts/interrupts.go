package interrupts

/**
 * Separate package exists mainly in order to avoid cyclic imports
 */

// Interrupt type - used to sygnalize incoming interrupt
// perhaps to be added:
// - delay
// - callback
// - callarg
type Interrupt struct {
	Priority  uint16
	Vector    uint16
	CleanFlag bool
}

// Trap is also a form of interrupt
type Trap struct {
	Vector uint16
	Msg    string
}

// interrupt vectors:

// TTYout : sent when a character is being printed on the teletype
const TTYout = 0064

// TTYin : sent when a key is punched on a teletype keyboard
const TTYin = 0060

/********************************
 * trap vectors:
 ********************************/

// IntBUS internal bus error?
const IntBUS = 04

// IntINVAL  - invalid (?)
const IntINVAL = 010

// IntDEBUG - debug trap
const TrapBRKPT = 014

// IntIOT - IO trap (?)
const IntIOT = 020

// IntFAULT - fault trap
const IntFAULT = 0250

// IntCLOCK : clock trap
const IntCLOCK = 0100

// IntRK - RK disk drive (?) interrupt
const IntRK = 0220

// InterruptQueue - to avoid keeping the insert to the queue login in unibus:
type InterruptQueue [8]Interrupt

// SendInterrupt to the queue
func (iq *InterruptQueue) SendInterrupt(priority, vector uint16) {
	interrupt := Interrupt{
		Priority: priority,
		Vector:   vector}

	if interrupt.Vector&1 == 1 {
		panic("Interrupt with Odd vector number")
	}

	// fast path:
	if iq[0].Vector == 0 {
		iq[0] = interrupt
		return
	}

	var i int
	for ; i < len(iq); i++ {
		if iq[i].Vector == 0 || iq[i].Priority < interrupt.Priority {
			break
		}
	}

	for ; i < len(iq); i++ {
		if iq[i].Vector == 0 || iq[i].Vector >= interrupt.Vector {
			break
		}
	}

	if i == len(iq) {
		panic("Interrupt table full")
	}

	for j := i + 1; j < len(iq); j++ {
		iq[j] = iq[j-1]
	}
	iq[i] = interrupt
}
