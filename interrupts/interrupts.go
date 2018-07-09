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

// interrupt vectors:

// IntTTYout : sent when character is being printed on the teletype
const IntTTYout = 0064
