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

// com - complement dst -> replace the contents of the destination address
// by their logical complement (each bit equal 0 is set to 1, each 1 is cleared)
func (c *CPU) comOp(instruction int16) error {
	dest := c.readWord(uint16(instruction & 077))
	c.writeWord(uint16(instruction&077), ^dest)
	return nil
}

// inc - increment dst
// todo: make sure overflow is set properly
// todo2: readWord should be able to handle reading operand with any addressing mode.
func (c *CPU) incOp(instruction int16) error {
	if addressMode := instruction & 070; addressMode == 0 {
		dst := c.Registers[instruction&7]
		result := dst + 1
		c.Registers[instruction&7] = result & 0xffff
		c.SetFlag("Z", result == 0)
		c.SetFlag("N", int16(result) < 0)
		c.SetFlag("V", dst == 0x7FFF)
	} else {
		dst := c.readWord(uint16(instruction & 077))
		res := dst + 1
		c.writeWord(uint16(instruction&077), res&0xffff)
		c.SetFlag("Z", res == 0)
		c.SetFlag("N", int16(res) < 0)
		c.SetFlag("V", dst == 0x7FFF)
	}
	return nil
}

// dec - decrement dst
func (c *CPU) decOp(instruction int16) error {
	if addressMode := instruction & 070; addressMode == 0 {
		dst := c.Registers[instruction&7]
		result := dst - 1
		c.Registers[instruction&7] = result & 0xffff // <- this is not really necessary, isn't it?
		c.SetFlag("Z", result == 0)
		c.SetFlag("N", int16(result) < 0)
		c.SetFlag("V", dst == 0x8000)
	} else {
		dst := c.readWord(uint16(instruction & 077))
		result := dst - 1
		c.writeWord(uint16(instruction&077), result&0xffff)
		c.SetFlag("Z", result == 0)
		c.SetFlag("N", int16(result) < 0)
		c.SetFlag("V", dst == 0x7FFF)
	}
	return nil
}

// neg - negate dst
// replace the contents of the destination address
// by it's 2 complement. 01000000 is replaced by itself
func (c *CPU) negOp(instruction int16) error {
	dest := c.readWord(uint16(instruction & 077))
	result := ^dest + 1
	c.writeWord(uint16(instruction&077), result)
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", int16(result) < 0)
	c.SetFlag("V", result == 0x8000)
	c.SetFlag("C", result != 0)
	return nil
}

// adc - add cary
func (c *CPU) adcOp(instruction int16) error {
	return nil
}

// sbc - substract carry
func (c *CPU) sbcOp(instruction int16) error {
	return nil
}

// tst - sets the condition cods N and Z according to the contents
// of the destination address
func (c *CPU) tstOp(instruction int16) error {
	dest := c.readWord(uint16(instruction & 077))

	if dest == 0 {
		c.SetFlag("Z", true)
	}
	if dest < 0 {
		c.SetFlag("N", true)
	}
	c.SetFlag("V", false)
	c.SetFlag("C", false)

	return nil
}

// asr - arithmetic shift right
func (c *CPU) asrOp(instruction int16) error {
	return nil
}

// asl - arithmetic shift left
func (c *CPU) aslOp(instruction int16) error {
	return nil
}

// ror - rotate right
func (c *CPU) rorOp(instruction int16) error {
	return nil
}

// rol - rorare left
func (c *CPU) rolOp(instruction int16) error {
	return nil
}

// jmp - jump to address:
func (c *CPU) jmpOp(instruction int16) error {
	return nil
}

// swab - swap bytes
func (c *CPU) swabOp(instruction int16) error {
	return nil
}

// mark - used as a part of subroutine return convention on pdp11
func (c *CPU) markOp(instruction int16) error {
	return nil
}

// mfpi - move from previous instruction space
func (c *CPU) mfpiOp(instruction int16) error {
	return nil
}

// mtpi - move to previous instruction space
func (c *CPU) mtpiOp(instruction int16) error {
	return nil
}

// sxt - sign extended
func (c *CPU) sxtOp(instruction int16) error {
	return nil
}

// double operand cpu instructions:

// move (1)
func (c *CPU) movOp(instruction int16) error {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	c.writeWord(uint16(dest), sourceVal)
	c.SetFlag("N", source < 0)
	c.SetFlag("Z", sourceVal == 0)
	// V is always cleared by MOV
	c.SetFlag("V", false)
	return nil
}

// misc instructions (decode all bits)
// halt
func (c *CPU) haltOp(instruction int16) error {
	// halt is an empty instruction. just stop CPU
	c.State = HALT
	return nil
}

// bpt - breakpoint trap
func (c *CPU) bptOp(instruction int16) error {
	return nil
}

// iot - i/o trap
func (c *CPU) iotOp(instruction int16) error {
	return nil
}

// rti - return from interrupt
func (c *CPU) rtiOp(instruction int16) error {
	return nil
}

// rtt - return from interrupt - same as rti, with distinction of inhibits a trace trap
func (c *CPU) rttOp(instruction int16) error {
	return nil
}

// wait
func (c *CPU) waitOp(instruction int16) error {
	return nil
}

func (c *CPU) resetOp(instruction int16) error {
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
	c.SetFlag("N", sum < 0)
	c.SetFlag("Z", sum == 0)
	if sourceVal > 0 && destVal > 0 && sum < 0 {
		c.SetFlag("V", true)
	}

	// this is possible, as type of sume is infered by compiler
	c.SetFlag("C", sum > 0xffff)

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

// jsr - jump to subroutine
func (c *CPU) jsrOp(instruction int16) error {
	return nil
}

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

// sob - substract one and branch (if not equal 0)
func (c *CPU) sobOp(instruction int16) error {
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

// trap opcodes:
// emt - emulator trap
func (c *CPU) emtOp(instruction int16) error {
	return nil
}

// trap
func (c *CPU) trapOp(instruction int16) error {
	return nil
}

// Single Register opcodes
// rts - return from subroutine
func (c *CPU) rtsOp(instruction int16) error {
	return nil
}

// clear flag opcodes
// covers following operations: CLN, CLZ, CLV, CLC, CCC
func (c *CPU) clearFlagOp(instruction int16) error {
	return nil
}

// set flag opcodes
// covers following operations: SEN, SEZ, SEV, SEC, SCC
func (c *CPU) setFlagOp(instruction int16) error {
	return nil
}
