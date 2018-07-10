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

// TTYout : sent when character is being printed on the teletype
const TTYout = 0064

// TTYin : sent when a key is punched on a teletype keyboard
const TTYin = 0060
