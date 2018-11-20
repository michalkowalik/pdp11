package unibus

// Definitions of the PDP CPU branching instructions

// branch calculates the branch to PC for a branch instruction offset
func (c *CPU) branch(instruction uint16) uint16 {

	// offset is an 8 bit signed integer
	var offset uint16
	var negBit bool
	pc := c.Registers[7]

	// offset is being kept in the low 8 bits of the command
	negBit = (instruction & 0200) > 0
	offset = uint16(instruction & 0xff)

	if negBit {
		offset = 0377 - offset + 1
	}

	if negBit {
		return pc - 2*offset
	}
	return pc + 2*offset
}

// control opcodes:
// br - unconditional branching (000400 + offset)
func (c *CPU) brOp(instruction uint16) error {
	c.Registers[7] = c.branch(instruction)
	return nil
}

// bne - branch if not equal (to zero) 0010000 + offset
func (c *CPU) bneOp(instruction uint16) error {
	if !c.GetFlag("Z") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// beq - branch if equal (to zero) 001400 + offset
func (c *CPU) beqOp(instruction uint16) error {
	if c.GetFlag("Z") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// bpl - branch if plus
func (c *CPU) bplOp(instruction uint16) error {
	if !c.GetFlag("N") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// bmi - branch if minus
func (c *CPU) bmiOp(instruction uint16) error {
	if c.GetFlag("N") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// bvc - branch if overflow is clear
func (c *CPU) bvcOp(instruction uint16) error {
	if !c.GetFlag("V") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// bvs - branch if overflow is set
func (c *CPU) bvsOp(instruction uint16) error {
	if c.GetFlag("V") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// bcc branch if carry is clear
func (c *CPU) bccOp(instruction uint16) error {
	if !c.GetFlag("C") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// bcs - branch if carry is set
func (c *CPU) bcsOp(instruction uint16) error {
	if c.GetFlag("C") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// bge - branch if greater than or equal (signed int)
func (c *CPU) bgeOp(instruction uint16) error {
	if c.GetFlag("N") == c.GetFlag("V") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// blt - branch if less than (zero)
func (c *CPU) bltOp(instruction uint16) error {
	if c.GetFlag("N") != c.GetFlag("V") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// bgt - branch if greater than (zero)
func (c *CPU) bgtOp(instruction uint16) error {
	if (c.GetFlag("V") == c.GetFlag("N")) && !c.GetFlag("Z") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// ble - branch if less than or equal
func (c *CPU) bleOp(instruction uint16) error {
	if (c.GetFlag("N") != c.GetFlag("V")) && !c.GetFlag("Z") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// bhi - branch if higher
func (c *CPU) bhiOp(instruction uint16) error {
	if !c.GetFlag("C") && !c.GetFlag("Z") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// blos - branch if lower or same
func (c *CPU) blosOp(instruction uint16) error {
	if c.GetFlag("C") != c.GetFlag("Z") {
		c.Registers[7] = c.branch(instruction)
	}
	return nil
}

// bhis - branch if higher or the same
func (c *CPU) bhisOp(instruction uint16) error {
	return c.bccOp(instruction)
}

// blo - branch if lower
func (c *CPU) bloOp(instruction uint16) error {
	return c.bccOp(instruction)
}
