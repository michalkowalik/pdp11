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
	vector    uint16
	cleanFlag bool
}
