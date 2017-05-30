package pdpcpu

// Definition of all PDP-11 CPU instructions
// All should follow the func (*CPU) (int16) signature

//getSignWord is useful to calculate the overflow bit:
func getSignWord(i uint16) uint16 {
	return (i >> 0xf) & 1
}

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

// move (1)
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

// compare (2)
func (c *CPU) cmpOp(instruction int16) error {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	destVal := c.readWord(uint16(dest))

	res := sourceVal - destVal

	c.SetFlag("N", res < 0)
	c.SetFlag("Z", res == 0)
	c.SetFlag("C", sourceVal < destVal)
	c.SetFlag("V", getSignWord((sourceVal^destVal)&(^destVal^res)) == 1)

	return nil
}

//add (6)
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

// substract (16)
func (c *CPU) subOp(instruction int16) error {
	return nil
}

//bit (3)
func (c *CPU) bitOp(instruction int16) error {
	return nil
}

// bit clear (4)
func (c *CPU) bicOp(instruction int16) error {
	return nil
}

// bit inclusive or (5)
func (c *CPU) bisOp(instruction int16) error {
	return nil
}

// RDD opcodes:
// multiply (070) --> EIS option, but let's have it
func (c *CPU) mulOp(instruction int16) error {
	return nil
}

// divide (071)
func (c *CPU) divOp(instruction int16) error {
	return nil
}

// shift arithmetically
func (c *CPU) ashOp(instruction int16) error {
	return nil
}

// arithmetic shift combined:
func (c *CPU) ashcOp(instruction int16) error {
	return nil
}

// xor
func (c *CPU) xorOp(instruction int16) error {
	return nil
}

// control opcodes:
// br - unconditional branching (000400 + offset)
func (c *CPU) brOp(instruction int16) error {
	return nil
}

// bne - branch if not equal (to zero) 0010000 + offset
func (c *CPU) bneOp(instruction int16) error {
	return nil
}

// beq - branch if equal (to zero) 001400 + offset
func (c *CPU) beqOp(instruction int16) error {
	return nil
}

// bpl - branch if plus
// bpl has rather weird opcode of 0100000
func (c *CPU) bplOp(instruction int16) error {
	return nil
}

// bmi - branch if minus
func (c *CPU) bmiOp(instruction int16) error {
	return nil
}

// bvc - branch if overflow is clear
func (c *CPU) bvcOp(instruction int16) error {
	return nil
}

// bvs - branch if overflow is set
func (c *CPU) bvsOp(instruction int16) error {
	return nil
}

// bcc branch if carry is clear
func (c *CPU) bccOp(instruction int16) error {
	return nil
}

// bcs - branch if carry is set
func (c *CPU) bcsOp(instruction int16) error {
	return nil
}
