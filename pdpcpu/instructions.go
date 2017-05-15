package pdpcpu

// Definition of all PDP-11 CPU instructions
// All should follow the func (*CPU) (int16) signature

// single operand cpu instructions:
func (c *CPU) clrOp(instruction int16) error {
	// address mode 0 -> clear register
	if addressMode := instruction & 0x38; addressMode == 0 {
		c.Registers[instruction&7] = 0
	} else {
		// TODO: access mode is hardcoded. needs to be changed or removed
		v, _ := c.mmunit.GetVirtualByMode(&c.Registers, uint16(instruction&0x3f), 1)
		c.mmunit.Memory[v] = 0
		c.mmunit.Memory[v+1] = 0
	}

	// set flags:
	c.SetFlag("C", false)
	c.SetFlag("N", false)
	c.SetFlag("V", false)
	c.SetFlag("Z", true)
	return nil
}

// double operand cpu instructions:
func (c *CPU) addOp(instruction int16) error {

	return nil
}
