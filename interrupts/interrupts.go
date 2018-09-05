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

/********************************
 * trap vectors:
 ********************************/

// INTBus internal bus error?
const INTBus = 04

// INTInval  - invalid (?)
const INTInval = 010

// INTDebug - debug trap
const INTDebug = 014

// INTIot - IO trap (?)
const INTIot = 020

// INTTtyIn - TTY trap
const INTTtyIn = 060

// INTFault - fault trap
const INTFault = 0250

// INTClock : clock trap
const INTClock = 0100

// INTRK - RK disk drive (?) trap
const INTRK = 0220
