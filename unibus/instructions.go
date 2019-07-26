package unibus

import (
	"pdp/psw"
)

// Definition of all PDP-11 CPU instructions
// All should follow the func (*CPU) (int16) signature

//getSignWord is useful to calculate the overflow bit:
func getSignWord(i uint16) uint16 {
	return (i >> 0xf) & 1
}

// single operand cpu instructions:
func (c *CPU) clrOp(instruction uint16) error {
	// address mode 0 -> clear register
	if addressMode := instruction & 0x38; addressMode == 0 {
		c.Registers[instruction&7] = 0
	} else {
		c.writeWord(instruction, 0)
	}

	// TODO: move all condition codes to psw kept by mmu
	// set flags:
	c.SetFlag("C", false)
	c.SetFlag("N", false)
	c.SetFlag("V", false)
	c.SetFlag("Z", true)
	return nil
}

// clr Byte
func (c *CPU) clrbOp(instruction uint16) error {
	return c.clrOp(instruction)
}

// com - complement dst -> replace the contents of the destination address
// by their logical complement (each bit equal 0 is set to 1, each 1 is cleared)
func (c *CPU) comOp(instruction uint16) error {
	dest := c.readWord(uint16(instruction & 077))
	c.writeWord(uint16(instruction&077), ^dest)
	return nil
}

func (c *CPU) combOp(instruction uint16) error {
	return c.comOp(instruction)
}

// inc - increment dst
// todo: make sure overflow is set properly
// todo2: readWord should be able to handle reading operand with any addressing mode.
func (c *CPU) incOp(instruction uint16) error {
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

func (c *CPU) incbOp(instruction uint16) error {
	return c.incOp(instruction)
}

// dec - decrement dst
func (c *CPU) decOp(instruction uint16) error {
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

func (c *CPU) decbOp(instruction uint16) error {
	return c.decOp(instruction)
}

// neg - negate dst
// replace the contents of the destination address
// by it's 2 complement. 01000000 is replaced by itself
func (c *CPU) negOp(instruction uint16) error {
	dest := c.readWord(instruction)
	result := ^dest + 1
	c.writeWord(instruction, result)
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", int16(result) < 0)
	c.SetFlag("V", result == 0x8000)
	c.SetFlag("C", result != 0)
	return nil
}

func (c *CPU) negbOp(instruction uint16) error {
	dest := c.readByte(instruction)
	result := ^dest + 1
	c.writeByte(instruction, uint16(result))
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", result&0x80 > 0)
	c.SetFlag("V", result == 0x80)
	c.SetFlag("C", result != 0)
	return nil
}

// adc - add cary
func (c *CPU) adcOp(instruction uint16) error {
	dest := c.readWord(uint16(instruction & 077))

	result := dest
	if c.GetFlag("C") {
		result = dest + 1
	}

	if err := c.writeWord(uint16(instruction&077), result); err != nil {
		return err
	}

	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("Z", result == 0)
	c.SetFlag("V", (dest == 077777) && c.GetFlag("C"))
	c.SetFlag("C", (dest == 0xFFFF) && c.GetFlag("C"))
	return nil
}

func (c *CPU) adcbOp(instruction uint16) error {
	return c.adcOp(instruction)
}

// sbc - substract carry
func (c *CPU) sbcOp(instruction uint16) error {
	dest := c.readWord(uint16(instruction & 077))
	result := dest
	if c.GetFlag("C") {
		result = result - 1
	}

	if err := c.writeWord(uint16(instruction&077), result); err != nil {
		return err
	}

	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("Z", result == 0)
	c.SetFlag("V", dest == 0x8000)
	c.SetFlag("V", (dest != 0) || c.GetFlag("C"))
	return nil
}

func (c *CPU) sbcbOp(instruction uint16) error {
	return c.sbcOp(instruction)
}

// tst - sets the condition cods N and Z according to the contents
// of the destination address
func (c *CPU) tstOp(instruction uint16) error {
	dest := c.readWord(uint16(instruction & 077))
	c.SetFlag("Z", dest == 0)
	c.SetFlag("N", (dest&0x8000) > 0)
	c.SetFlag("V", false)
	c.SetFlag("C", false)
	return nil
}

func (c *CPU) tstbOp(instruction uint16) error {
	dest := c.readByte(uint16(instruction & 077))
	c.SetFlag("Z", dest == 0)
	c.SetFlag("N", (dest&0x80) > 0)
	c.SetFlag("V", false)
	c.SetFlag("C", false)
	return nil
}

// asr - arithmetic shift right
// 	Shifts all bits of the destination right one place. Bit 15
// is replicated. The C-bit is loaded from bit 0 of the destination.
// ASR performs signed division of the destination by two.
func (c *CPU) asrOp(instruction uint16) error {
	dest := c.readWord(uint16(instruction & 077))
	result := (dest & 0x8000) | (dest >> 1)
	if err := c.writeWord(uint16(instruction&077), result); err != nil {
		return err
	}
	c.SetFlag("C", (dest&1) == 1)
	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("Z", result == 0)

	// V flag is a XOR on C and N flag, but golang doesn't provide boolean XOR
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
	return nil
}

func (c *CPU) asrbOp(instruction uint16) error {
	return c.asrOp(instruction)
}

// asl - arithmetic shift left
// Shifts all bits of the destination left one place. Bit 0 is
// loaded with an 0. The C·bit of the status word is loaded from
// the most significant bit of the destination. ASL performs a
// signed multiplication of the destination by 2 with overflow indication.
func (c *CPU) aslOp(instruction uint16) error {
	dest := c.readWord(uint16(instruction & 077))
	result := dest << 1
	if err := c.writeWord(uint16(instruction&077), result); err != nil {
		return err
	}
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("C", (dest&0x8000) == 0x8000)
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
	return nil
}

func (c *CPU) aslbOp(instruction uint16) error {
	return c.aslOp(instruction)
}

// ror - rotate right
// Rotates all bits of the destination right one place. Bit 0 is
// loaded into the C-bit and the previous contents of the C-bit
// are loaded into bit 15 of the destination.
func (c *CPU) rorOp(instruction uint16) error {
	dest := c.readWord(uint16(instruction & 077))
	cBit := (dest & 1) << 15
	result := (dest >> 1) | cBit
	if err := c.writeWord(uint16(instruction&077), result); err != nil {
		return err
	}
	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("Z", result == 0)
	c.SetFlag("C", cBit > 0)
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
	return nil
}

func (c *CPU) rorbOp(instruction uint16) error {
	return c.rorOp(instruction)
}

// rol - rorare left
// : Rotate all bits of the destination left one place. Bit 15
// is loaded into the C·bit of the status word and the previous
// contents of the C-bit are loaded into Bit 0 of the destination.
func (c *CPU) rolOp(instruction uint16) error {
	dest := c.readWord(uint16(instruction & 077))
	c.SetFlag("C", (dest&0x8000) == 0x8000)
	result := dest >> 1
	if c.GetFlag("C") {
		result = result | 1
	}
	if err := c.writeWord(uint16(instruction&077), result); err != nil {
		return err
	}
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
	return nil
}

func (c *CPU) rolbOp(instruction uint16) error {
	return c.rolOp(instruction)
}

// jmp - jump to address:
func (c *CPU) jmpOp(instruction uint16) error {
	return nil
}

// swab - swap bytes
// Exchanges high-order byte and low-order byte of the destination
// word (destination must be a word address).
// N: set if high-order bit of low-order byte (bit 7) of result is set;
// cleared otherwise
// Z: set if low-order byte of result = 0; cleared otherwise
// V: cleared
// C: cleared
func (c *CPU) swabOp(instruction uint16) error {
	dest := c.readWord(uint16(instruction & 077))
	result := (dest << 8) | (dest >> 8)

	if err := c.writeWord(uint16(instruction&077), result); err != nil {
		return err
	}

	c.SetFlag("N", (result&0x80) == 0x80)
	c.SetFlag("Z", (result&0xff) == 0)
	c.SetFlag("V", false)
	c.SetFlag("C", false)
	return nil
}

// mark - used as a part of subroutine return convention on pdp11
func (c *CPU) markOp(instruction uint16) error {
	return nil
}

// mfpi - move from previous instruction space
func (c *CPU) mfpiOp(instruction uint16) error {
	return nil
}

// mtpi - move to previous instruction space
func (c *CPU) mtpiOp(instruction uint16) error {
	return nil
}

// sxt - sign extended
// If the condition code bit N is set then a -1 is placed in the
// destination operand: if N bit is clear, then a 0 is placed in the
// destination operand.
func (c *CPU) sxtOp(instruction uint16) error {
	res := 0
	if c.GetFlag("N") {
		res = -1
	}

	if err := c.writeWord(uint16(instruction&077), uint16(res)); err != nil {
		return err
	}

	c.SetFlag("Z", !c.GetFlag("N"))
	return nil
}

// double operand cpu instructions:

// move (1)
func (c *CPU) movOp(instruction uint16) error {
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

// movb
// TODO: Finish implementation
func (c *CPU) movbOp(instruction uint16) error {
	return c.movOp(instruction)
}

// misc instructions (decode all bits)
// halt
func (c *CPU) haltOp(instruction uint16) error {
	// halt is an empty instruction. just stop CPU
	c.State = HALT
	return nil
}

// bpt - breakpoint trap
func (c *CPU) bptOp(instruction uint16) error {
	// 14 is breakpoint trap vector
	c.Trap(014)
	return nil
}

// iot - i/o trap
func (c *CPU) iotOp(instruction uint16) error {
	c.Trap(020)
	return nil
}

// rti - return from interrupt
func (c *CPU) rtiOp(instruction uint16) error {
	c.Registers[7] = c.Pop()
	tempPsw := psw.PSW(c.Pop())
	// TODO: add the psw modification if cpu in user mode

	c.mmunit.Psw = &tempPsw
	return nil
}

// rtt - return from trap
func (c *CPU) rttOp(instruction uint16) error {
	return c.rtiOp(instruction)
}

// wait for interrupt
// check for interrupts here!!
func (c *CPU) waitOp(instruction uint16) error {
	c.State = WAIT

	// update display:

	return nil
}

// Sends INIT on UNIBUS for 10ms. All devices on the UNIBUS are reset and power up
// Implementation needs to wait for the unibus.
func (c *CPU) resetOp(instruction uint16) error {
	c.Reset()
	return nil
}

// compare (2)
func (c *CPU) cmpOp(instruction uint16) error {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	destVal := c.readWord(uint16(dest))

	res := int(sourceVal) - int(destVal)

	c.SetFlag("N", res < 0)
	c.SetFlag("Z", res == 0)
	c.SetFlag("C", sourceVal < destVal)
	c.SetFlag("V", getSignWord((sourceVal^destVal)&(^destVal^uint16(res))) == 1)

	return nil
}

func (c *CPU) cmpbOp(instruction uint16) error {
	return c.cmpOp(instruction)
}

//add (6)
func (c *CPU) addOp(instruction uint16) error {
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

	// this is possible, as type of sum is infered by compiler
	c.SetFlag("C", sum > 0xffff)

	c.writeWord(uint16(dest), uint16(sum)&0xffff)
	return nil
}

// substract (16)
func (c *CPU) subOp(instruction uint16) error {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	destVal := c.readWord(uint16(dest))

	res := destVal - sourceVal
	c.SetFlag("C", sourceVal > destVal)
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", res < 0)
	c.SetFlag("V", getSignWord((sourceVal^destVal)&(^destVal^res)) == 1)

	c.writeWord(uint16(dest), uint16(res)&0xffff)
	return nil
}

//bit (3)
func (c *CPU) bitOp(instruction uint16) error {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	destVal := c.readWord(uint16(dest))

	res := sourceVal & destVal
	c.SetFlag("V", false)
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", (res&0x8000) > 0)
	return nil
}

func (c *CPU) bitbOp(instruction uint16) error {
	return c.bitOp(instruction)
}

// bit clear (4)
func (c *CPU) bicOp(instruction uint16) error {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	destVal := c.readWord(uint16(dest))

	destVal = destVal & (^sourceVal)
	c.SetFlag("V", false)
	c.SetFlag("N", (destVal&0x8000) > 0)
	c.SetFlag("Z", destVal == 0)
	c.writeWord(uint16(dest), uint16(destVal)&0xffff)
	return nil
}

func (c *CPU) bicbOp(instruction uint16) error {
	return c.bicOp(instruction)
}

// bit inclusive or (5)
func (c *CPU) bisOp(instruction uint16) error {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	destVal := c.readWord(uint16(dest))

	destVal = destVal | sourceVal
	c.SetFlag("V", false)
	c.SetFlag("N", (destVal&0x8000) > 0)
	c.SetFlag("Z", destVal == 0)
	c.writeWord(uint16(dest), uint16(destVal)&0xffff)
	return nil
}

func (c *CPU) bisbOp(instruction uint16) error {
	return c.bisOp(instruction)
}

// RDD opcodes:

// jsr - jump to subroutine
func (c *CPU) jsrOp(instruction uint16) error {
	register := (instruction >> 6) & 7
	destination := uint16(instruction & 077)
	c.Push(uint16(c.Registers[register]))
	c.Registers[register] = c.Registers[7]
	c.Registers[7] = c.readWord(destination)

	// tmp only:
	// c.mmunit.DumpMemory()

	return nil
}

// multiply (070) --> EIS option, but let's have it
func (c *CPU) mulOp(instruction uint16) error {
	return nil
}

// divide (071)
func (c *CPU) divOp(instruction uint16) error {
	return nil
}

// shift arithmetically
func (c *CPU) ashOp(instruction uint16) error {

	register := (instruction >> 6) & 7
	offset := uint16(instruction & 077)
	if offset == 0 {
		return nil
	}
	result := uint16(c.Registers[register])

	// negative number -> shift right
	if (offset & 040) > 0 {
		offset = 64 - offset
		if offset > 16 {
			offset = 16
		}

		// set C flag:
		c.SetFlag("C", (result&(0x8000>>(16-offset))) != 0)
		result = result >> offset
	} else {
		if offset > 16 {
			result = 0
		} else {
			// set C flag:
			c.SetFlag("C", (result&(1<<(16-offset))) != 0)
			result = result << offset
		}
	}

	// V flag set if sign changed:
	c.SetFlag("V", (c.Registers[register]&0x8000) != (result&0x8000))
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", result < 0)
	c.Registers[register] = result
	return nil
}

// arithmetic shift combined (EIS option)
func (c *CPU) ashcOp(instruction uint16) error {

	var result uint32
	offset := uint16(instruction & 077)
	if offset == 0 {
		return nil
	}

	register := (instruction >> 6) & 7
	dst := (uint32(c.Registers[register]) << 16) | uint32(c.Registers[register|1])

	// negative number -> shift right
	if (offset & 040) > 0 {
		offset = 64 - offset
		if offset > 32 {
			offset = 32
		}
		result = dst >> (offset - 1)
		c.SetFlag("C", (result&0x0001) != 0)
		result = result >> 1
		if (dst & 0x80000000) != 0 {
			// TODO: Why???
			result = result | (0xffffffff << (32 - offset))
		}
	} else {
		result = dst << (offset - 1)
		c.SetFlag("C", (result&0x8000) != 0)
		result = result << 1

	}

	c.Registers[register] = uint16((result >> 16) & 0xffff)
	c.Registers[register|1] = uint16(result & 0xffff)
	c.SetFlag("N", result < 0)
	c.SetFlag("Z", result == 0)
	// V flag set if the sign bit changed during the shift
	c.SetFlag("V", (dst>>31) != (result>>31))

	return nil
}

// xor
func (c *CPU) xorOp(instruction uint16) error {
	sourceVal := c.Registers[(instruction>>6)&7]
	dest := instruction & 077
	destVal := c.readWord(uint16(dest))

	res := sourceVal ^ destVal

	c.SetFlag("N", res < 0)
	c.SetFlag("Z", res == 0)
	c.SetFlag("V", false)

	c.writeWord(uint16(dest), uint16(res))
	return nil
}

// sob - substract one and branch (if not equal 0)
// if value of the register sourceReg is not 0, susbtract
// twice the value of the offset (lowest 6 bits) from the SP
func (c *CPU) sobOp(instruction uint16) error {
	sourceReg := (instruction >> 6) & 7
	c.Registers[sourceReg] = (c.Registers[sourceReg] - 1) & 0xffff
	if c.Registers[sourceReg] != 0 {
		c.Registers[7] = (c.Registers[7] - ((uint16(instruction) & 077) << 1)) & 0xffff
	}
	return nil
}

// trap opcodes:
// emt - emulator trap - trap vector hardcoded to location 32
func (c *CPU) emtOp(instruction uint16) error {
	c.Trap(32)
	return nil
}

// trap
// trap vector for TRAP is hardcoded for all PDP11s to memory location 34
func (c *CPU) trapOp(instruction uint16) error {
	c.Trap(34)
	return nil
}

// Single Register opcodes
// rts - return from subroutine
func (c *CPU) rtsOp(instruction uint16) error {
	register := instruction & 7

	// load Program Counter from register passed in instruction
	c.Registers[7] = c.Registers[register]

	// load word popped from processor stack to "register"
	c.Registers[register] = c.Pop()
	return nil
}

// clear flag opcodes
// covers following operations: CLN, CLZ, CLV, CLC, CCC
func (c *CPU) clearFlagOp(instruction uint16) error {

	switch flag := instruction & 0777; flag {
	case 0241:
		c.SetFlag("C", false)
	case 0242:
		c.SetFlag("V", false)
	case 0243:
		c.SetFlag("C", false)
		c.SetFlag("V", false)
	case 0244:
		c.SetFlag("Z", false)
	case 0250:
		c.SetFlag("N", false)
	case 0257:
		c.SetFlag("N", false)
		c.SetFlag("Z", false)
		c.SetFlag("C", false)
		c.SetFlag("V", false)
	}
	return nil
}

// set flag opcodes
// covers following operations: SEN, SEZ, SEV, SEC, SCC
func (c *CPU) setFlagOp(instruction uint16) error {
	switch flag := instruction & 0777; flag {
	case 0261:
		c.SetFlag("C", true)
	case 0262:
		c.SetFlag("V", true)
	case 0264:
		c.SetFlag("Z", true)
	case 0270:
		c.SetFlag("N", true)
	case 0277:
		c.SetFlag("N", true)
		c.SetFlag("Z", true)
		c.SetFlag("C", true)
		c.SetFlag("V", true)
	}
	return nil
}
