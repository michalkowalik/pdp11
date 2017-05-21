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
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	destVal := c.readWord(uint16(dest))

	sum := sourceVal + destVal
	if sum < 0 {
		c.SetFlag("N", true)
	}
	if sum == 0 {
		c.SetFlag("Z", true)
	}
	if sourceVal > 0 && destVal > 0 && sum < 0 {
		c.SetFlag("V", true)
	}
	if sum > 0xffff {
		c.SetFlag("C", true)
	}
	c.writeWord(uint16(dest), uint16(sum)&0xffff)
	return nil
}

func (c *CPU) movOp(instruction int16) error {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	c.writeWord(uint16(dest), sourceVal)
	if sourceVal < 0 {
		c.SetFlag("N", true)
	}
	if sourceVal == 0 {
		c.SetFlag("Z", true)
	}
	// V is always cleared by MOV
	c.SetFlag("V", false)
	return nil
}

func (c *CPU) haltOp(instruction int16) error {
	// halt is an empty instruction. just stop CPU
	c.State = HALT
	return nil
}
