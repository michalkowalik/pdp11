package pdpcpu

// Definitions of the PDP CPU branching instructions

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

// bge - branch if greater than or equal (signed int)
func (c *CPU) bgeOp(instruction int16) error {
	return nil
}

// blt - branch if less than (zero)
func (c *CPU) bltOp(instruction int16) error {
	return nil
}

// bgt - branch if greater than (zero)
func (c *CPU) bgtOp(instruction int16) error {
	return nil
}

// ble - branch if less than or equal
func (c *CPU) bleOp(instruction int16) error {
	return nil
}

// bhi - branch if higher
func (c *CPU) bhiOp(instruction int16) error {
	return nil
}

// blos - branch if lower or same
func (c *CPU) blosOp(instruction int16) error {
	return nil
}

// bhis - branch if higher or the same
func (c *CPU) bhisOp(instruction int16) error {
	return nil
}

// blo - branch if lower
func (c *CPU) bloOp(instruction int16) error {
	return nil
}
