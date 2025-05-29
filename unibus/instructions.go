package unibus

import (
	"pdp/psw"
)

// Definition of all PDP-11 CPU instructions
// All should follow the func (*CPU) (int16) signature

// single operand cpu instructions:
// clr Word / Byte
func (c *CPU) clrOp(instruction uint16) {
	if instruction&0100000 == 0100000 {
		dstAddr := c.GetVirtualAddress(instruction&077, 1)
		c.mmunit.WriteMemoryByte(dstAddr, 0)
	} else {
		dstAddr := c.GetVirtualAddress(instruction&077, 0)
		c.mmunit.WriteMemoryWord(dstAddr, 0)
	}
	c.SetFlag("C", false)
	c.SetFlag("N", false)
	c.SetFlag("V", false)
	c.SetFlag("Z", true)
}

// com - complement dst -> replace the contents of the destination address
// by their logical complement (each bit equal 0 is set to 1, each 1 is cleared)
func (c *CPU) comOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 0)
	dst := c.mmunit.ReadMemoryWord(dstAddr)

	c.SetFlag("N", dst&0x8000 == 0x8000)
	c.SetFlag("Z", dst == 0)
	c.SetFlag("C", true)
	c.SetFlag("V", false)
	c.mmunit.WriteMemoryWord(dstAddr, ^dst)
}

func (c *CPU) combOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 0)
	dst := c.mmunit.ReadMemoryByte(dstAddr)

	c.SetFlag("N", dst&0x80 == 0x80)
	c.SetFlag("Z", dst == 0)
	c.SetFlag("C", true)
	c.SetFlag("V", false)
	c.mmunit.WriteMemoryByte(dstAddr, ^dst)
}

// inc - increment dst
func (c *CPU) incOp(instruction uint16) {
	dest := instruction & 077
	virtAddr := c.GetVirtualAddress(dest, 0)
	val := (c.mmunit.ReadMemoryWord(virtAddr) + 1) & 0xFFFF
	c.mmunit.WriteMemoryWord(virtAddr, val)

	c.SetFlag("Z", val == 0)
	c.SetFlag("N", val&0x8000 == 0x8000)
	c.SetFlag("V", val&0x8000 == 0x8000)
}

func (c *CPU) incbOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 1)
	val := c.mmunit.ReadMemoryByte(dstAddr)
	res := (val + 1) & 0xFF
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", res&0x80 == 0x80)
	c.SetFlag("V", val == 0x7F)
	c.mmunit.WriteMemoryByte(dstAddr, res)
}

// dec - decrement dst
// TODO: it should look like INC
func (c *CPU) decOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 0)
	val := c.mmunit.ReadMemoryWord(dstAddr)
	res := (val - 1) & 0xFFFF

	c.SetFlag("Z", res == 0)
	c.SetFlag("N", res&0x8000 == 0x8000)
	c.SetFlag("V", val == 0100000)
	c.mmunit.WriteMemoryWord(dstAddr, res)
}

func (c *CPU) decbOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 1)
	val := c.mmunit.ReadMemoryByte(dstAddr)
	res := (val - 1) & 0xFF

	c.SetFlag("Z", res == 0)
	c.SetFlag("N", res&0x80 == 0x80)
	c.SetFlag("V", val == 0x80)
	c.mmunit.WriteMemoryByte(dstAddr, res)
}

// neg - negate dst
// replace the contents of the destination address
// by its 2 complement. 01000000 is replaced by itself
func (c *CPU) negOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(dstAddr)
	result := ^dest + 1
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", int16(result) < 0)
	c.SetFlag("V", result == 0x8000)
	c.SetFlag("C", result != 0)
	c.mmunit.WriteMemoryWord(dstAddr, result)
}

func (c *CPU) negbOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(dstAddr)
	result := ^dest + 1
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", result&0x80 > 0)
	c.SetFlag("V", result == 0x80)
	c.SetFlag("C", result != 0)
	c.mmunit.WriteMemoryByte(dstAddr, result)
}

// adc - add cary
func (c *CPU) adcOp(instruction uint16) {
	var dstAddr uint16
	var dst uint16
	var msb uint16 = 0x8000
	var ov uint16 = 077777
	var oc uint16 = 0xFFFF

	if instruction&0100000 == 0100000 {
		dstAddr = c.GetVirtualAddress(instruction&077, 1)
		dst = uint16(c.mmunit.ReadMemoryByte(dstAddr))
		msb = 0x80
		ov = 0177
		oc = 0xFF
	} else {
		dstAddr = c.GetVirtualAddress(instruction&077, 0)
		dst = c.mmunit.ReadMemoryWord(dstAddr)

	}
	result := dst
	if c.GetFlag("C") {
		result = dst + 1
		if instruction&0100000 == 0100000 {
			c.mmunit.WriteMemoryByte(dstAddr, byte(result&oc))
		} else {
			c.mmunit.WriteMemoryWord(dstAddr, result)
		}
	}
	c.SetFlag("N", (result&msb) == msb)
	c.SetFlag("Z", result == 0)
	c.SetFlag("V", (dst == ov) && c.GetFlag("C"))
	c.SetFlag("C", (dst == oc) && c.GetFlag("C"))
}

// sbc - subtract carry
func (c *CPU) sbcOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(dstAddr)
	result := dest
	if c.GetFlag("C") {
		result = result - 1
		c.mmunit.WriteMemoryWord(dstAddr, result)
	}

	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("Z", result == 0)
	c.SetFlag("V", dest == 0x8000)
	c.SetFlag("C", !((dest == 0) && c.GetFlag("C")))
}

func (c *CPU) sbcbOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(dstAddr)
	result := dest
	if c.GetFlag("C") {
		result = result - 1
		c.mmunit.WriteMemoryByte(dstAddr, result)
	}

	c.SetFlag("N", (result&0x80) == 0x80)
	c.SetFlag("Z", result == 0)
	c.SetFlag("V", dest == 0x80)
	c.SetFlag("C", !((dest == 0) && c.GetFlag("C")))
}

// tst - sets the condition codes N and Z according to the contents
// of the destination address
func (c *CPU) tstOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(dstAddr)

	c.SetFlag("Z", dest == 0)
	c.SetFlag("N", (dest&0x8000) > 0)
	c.SetFlag("V", false)
	c.SetFlag("C", false)
}

func (c *CPU) tstbOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(dstAddr)

	c.SetFlag("Z", dest == 0)
	c.SetFlag("N", (dest&0x80) > 0)
	c.SetFlag("V", false)
	c.SetFlag("C", false)
}

// asr - arithmetic shift right
//
//	Shifts all bits of the destination right one place. Bit 15
//
// is replicated. The C-bit is loaded from bit 0 of the destination.
// ASR performs signed division of the destination by two.
func (c *CPU) asrOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(dstAddr)
	result := (dest & 0x8000) | (dest >> 1)
	c.mmunit.WriteMemoryWord(dstAddr, result)

	c.SetFlag("C", (dest&1) == 1)
	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("Z", result == 0)

	// V flag is a XOR on C and N flag, but golang doesn't provide boolean XOR
	c.SetFlag("V", c.GetFlag("C") != c.GetFlag("N"))
}

func (c *CPU) asrbOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(dstAddr)
	result := (dest & 0x80) | (dest >> 1)
	c.mmunit.WriteMemoryByte(dstAddr, result)

	c.SetFlag("C", (dest&1) == 1)
	c.SetFlag("N", (result&0x80) == 0x80)
	c.SetFlag("Z", result == 0)

	// V flag is a XOR on C and N flag, but golang doesn't provide boolean XOR
	c.SetFlag("V", c.GetFlag("C") != c.GetFlag("N"))
}

// asl - arithmetic shift left
// Shifts all bits of the destination left one place. Bit 0 is
// loaded with an 0. The C·bit of the status word is loaded from
// the most significant bit of the destination. ASL performs a
// signed multiplication of the destination by 2 with overflow indication.
func (c *CPU) aslOp(instruction uint16) {
	destAddr := c.GetVirtualAddress(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(destAddr)
	result := dest << 1
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("C", (dest&0x8000) == 0x8000)
	c.SetFlag("V", c.GetFlag("C") != c.GetFlag("N"))
	c.mmunit.WriteMemoryWord(destAddr, result)
}

func (c *CPU) aslbOp(instruction uint16) {
	destAddr := c.GetVirtualAddress(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(destAddr)
	result := dest << 1
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", (result&0x80) == 0x80)
	c.SetFlag("C", (dest&0x80) == 0x80)
	c.SetFlag("V", c.GetFlag("C") != c.GetFlag("N"))
	c.mmunit.WriteMemoryByte(destAddr, result)
}

// ror - rotate right
// Rotates all bits of the destination right one place. Bit 0 is
// loaded into the C-bit and the previous contents of the C-bit
// are loaded into bit 15 of the destination.
func (c *CPU) rorOp(instruction uint16) {
	destAddr := c.GetVirtualAddress(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(destAddr)
	cBit := (dest & 1) << 15
	result := (dest >> 1) | cBit
	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("Z", result == 0)
	c.SetFlag("C", cBit > 0)
	c.SetFlag("V", c.GetFlag("C") != c.GetFlag("N"))
	c.mmunit.WriteMemoryWord(destAddr, result)
}

func (c *CPU) rorbOp(instruction uint16) {
	destAddr := c.GetVirtualAddress(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(destAddr)
	cBit := (dest & 1) << 7
	result := (dest >> 1) | cBit
	c.SetFlag("N", (result&0x80) == 0x80)
	c.SetFlag("Z", result == 0)
	c.SetFlag("C", cBit > 0)
	c.SetFlag("V", c.GetFlag("C") != c.GetFlag("N"))
	c.mmunit.WriteMemoryByte(destAddr, result)
}

// rol - rorare left
// : Rotate all bits of the destination left one place. Bit 15
// is loaded into the C·bit of the status word and the previous
// contents of the C-bit are loaded into Bit 0 of the destination.
func (c *CPU) rolOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(dstAddr)
	res := dest << 1

	if c.GetFlag("C") {
		res |= 1
	}
	c.SetFlag("C", (dest&0x8000) == 0x8000)
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", (res&0x8000) == 0x8000)
	c.SetFlag("V", c.GetFlag("C") != c.GetFlag("N"))
	c.mmunit.WriteMemoryWord(dstAddr, res)
}

func (c *CPU) rolbOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(dstAddr)
	res := dest << 1

	if c.GetFlag("C") {
		res |= 1
	}
	c.SetFlag("C", (dest&0x80) == 0x80)
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", (res&0x80) == 0x80)
	c.SetFlag("V", c.GetFlag("C") != c.GetFlag("N"))
	c.mmunit.WriteMemoryByte(dstAddr, res)
}

// jmp - jump to address:
func (c *CPU) jmpOp(instruction uint16) {
	dest := c.GetVirtualAddress(instruction&077, 0) // c.readWord(uint16(instruction & 077))
	if instruction&070 == 0 {
		panic("JMP: Can't jump to register")
	}
	c.Registers[7] = dest
}

// swab - swap bytes
// Exchanges high-order byte and low-order byte of the destination
// word (destination must be a word address).
// N: set if high-order bit of low-order byte (bit 7) of result is set;
// cleared otherwise
// Z: set if low-order byte of result = 0; cleared otherwise
// V: cleared
// C: cleared
func (c *CPU) swabOp(instruction uint16) {
	dstAddr := c.GetVirtualAddress(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(dstAddr)
	result := (dest << 8) | (dest >> 8)

	c.SetFlag("N", (result&0x80) == 0x80)
	c.SetFlag("Z", (result&0xff) == 0)
	c.SetFlag("V", false)
	c.SetFlag("C", false)
	c.mmunit.WriteMemoryWord(dstAddr, result)
}

// mark - used as a part of subroutine return convention on pdp11
func (c *CPU) markOp(instruction uint16) {
	c.Registers[6] = c.Registers[7] + (instruction&0xFFFF)<<1
	c.Registers[7] = c.Registers[5]
	c.Registers[5] = c.Pop()
}

// mfpi - move from previous instruction space
func (c *CPU) mfpiOp(instruction uint16) {
	var val uint16
	dest := c.GetVirtualAddress(instruction&077, 0)

	switch {
	case dest == 0177706:
		if c.IsUserMode() == c.IsPrevModeUser() {
			val = c.Registers[6]
		} else {
			if c.IsPrevModeUser() { // user
				val = c.UserStackPointer
			} else {
				val = c.KernelStackPointer
			}
		}
	// if register:
	case dest&0177770 == 0170000:
		panic("MFPI attended on Register address")
	default:
		physicalAddress := c.mmunit.Decode(dest, false, c.IsPrevModeUser())
		val = c.unibus.ReadIO(physicalAddress)
	}

	c.Push(val)
	c.SetFlag("C", true)
	c.SetFlag("N", val&0x8000 == 0x8000)
	c.SetFlag("Z", val == 0)
}

// mtpi - move to previous instruction space
func (c *CPU) mtpiOp(instruction uint16) {
	destAddr := c.GetVirtualAddress(instruction&077, 0)
	val := c.Pop()

	switch {
	case destAddr == 0177706:
		if c.IsUserMode() == c.IsPrevModeUser() {
			c.Registers[6] = val
		} else {
			if c.IsPrevModeUser() {
				c.UserStackPointer = val
			} else {
				c.KernelStackPointer = val
			}
		}
	case destAddr&0177770 == 0170000:
		panic("MTPI attended on Register address")
	default:
		sourceAddress := c.mmunit.Decode(destAddr, false, c.IsPrevModeUser())
		c.unibus.WriteIO(sourceAddress, val)
	}

	// TODO: Not strictly needed
	c.unibus.Psw.Set(c.unibus.Psw.Get() & 0xFFFF)
	c.SetFlag("C", true)
	c.SetFlag("N", val&0x8000 == 0x8000)
	c.SetFlag("Z", val == 0)
}

// sxt - sign extended
// If the condition code bit N is set then a -1 is placed in the
// destination operand: if N bit is clear, then a 0 is placed in the
// destination operand.
func (c *CPU) sxtOp(instruction uint16) {
	res := 0
	if c.GetFlag("N") {
		res = -1
	}
	c.writeWord(instruction&077, uint16(res))
	c.SetFlag("Z", !c.GetFlag("N"))
}

// double operand cpu instructions:
// move (1)
func (c *CPU) movOp(instruction uint16) {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	srcAddr := c.GetVirtualAddress(source, 0)
	sourceVal := c.mmunit.ReadMemoryWord(srcAddr)
	dstAddr := c.GetVirtualAddress(dest, 0)

	c.SetFlag("N", (sourceVal&0x8000) > 0)
	c.SetFlag("Z", sourceVal == 0)
	// V is always cleared by MOV
	c.SetFlag("V", false)
	c.mmunit.WriteMemoryWord(dstAddr, sourceVal)
}

// movb
// The MOVB to a register (unique among byte instructions)
// extends the most significant bit of the low order byte (sign extension).
// Otherwise, MOVB operates on bytes exactly as MOV operates on words.
func (c *CPU) movbOp(instruction uint16) {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceAddr := c.GetVirtualAddress(source, 1)
	sourceVal := c.mmunit.ReadMemoryByte(sourceAddr)
	destAddr := c.GetVirtualAddress(dest, 1)

	c.SetFlag("Z", sourceVal == 0)
	c.SetFlag("V", false)
	c.SetFlag("N", (sourceVal&0x80) > 0)

	// register destination is a special case in movb:
	if dest&070 == 0 {
		if sourceVal&0x80 == 0x80 {
			c.Registers[dest&7] = uint16(0xFF00) | uint16(sourceVal)
			return
		}
	}
	c.mmunit.WriteMemoryByte(destAddr, sourceVal)
}

// misc instructions (decode all bits)
// halt
func (c *CPU) haltOp(_ uint16) {
	c.State = HALT
}

// bpt - breakpoint trap
func (c *CPU) bptOp(_ uint16) {
	c.trapOpcode(014)
}

// iot - i/o trap
func (c *CPU) iotOp(_ uint16) {
	c.trapOpcode(020)
}

// rti - return from interrupt
func (c *CPU) rtiOp(_ uint16) {
	c.log.Printf("calling rti \n")
	// DEBUG: POP from interrupt stack
	//c.unibus.InterruptStack.Pop()

	c.Registers[7] = c.Pop()
	val := c.Pop()      // pop the PSW
	if c.IsUserMode() { // why does it happen at all?
		// DEBUG code
		c.log.Printf("interrupt return in user mode\n")

		// why is that needed at all??
		val &= 047                          // Save the flags
		val |= c.unibus.Psw.Get() & 0177730 // how is that correct?
	}
	c.unibus.WriteIO(PSWAddr, val)
}

// rtt - return from trap
func (c *CPU) rttOp(_ uint16) {
	//c.unibus.InterruptStack.Pop()

	c.Registers[7] = c.Pop()
	val := c.Pop()      // pop the PSW
	if c.IsUserMode() { // why does it happen at all?
		c.log.Printf("Trap return in user mode\n")
		val &= 047                          // Save the flags
		val |= c.unibus.Psw.Get() & 0177730 // how is that correct?
	}
	c.unibus.WriteIO(PSWAddr, val)
	// c.rtiOp(instruction)
}

// wait for interrupt
func (c *CPU) waitOp(_ uint16) {
	if !c.IsUserMode() {
		c.State = WAIT
	}
}

// Sends INIT on UNIBUS for 10ms. All devices on the UNIBUS are reset and power up
func (c *CPU) resetOp(_ uint16) {
	c.unibus.Rk01.Reset()
	c.unibus.TermEmulator.ClearTerminal()
}

// compare (2) - byte op included
func (c *CPU) cmpOp(instruction uint16) {
	byteOp := instruction&0100000 > 0
	source := (instruction & 07700) >> 6
	dest := instruction & 077
	msb := uint16(0100000)

	var sourceVal uint16
	var destVal uint16

	if byteOp {
		msb = 0200
		sourceVal = uint16(c.readByte(source))
		destVal = uint16(c.readByte(dest))
	} else {
		sourceVal = c.readWord(source)
		destVal = c.readWord(dest)
	}
	res := sourceVal + (^(destVal) + 1)

	c.SetFlag("N", (res&msb) > 0)
	c.SetFlag("Z", res == 0)
	c.SetFlag("C", sourceVal < destVal)
	c.SetFlag("V", (sourceVal^destVal)&msb == msb && !((destVal^res)&msb == msb))
}

// add (6)
func (c *CPU) addOp(instruction uint16) {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(source)
	virtAddr := c.GetVirtualAddress(dest, 0)
	destVal := c.mmunit.ReadMemoryWord(virtAddr)
	sum := sourceVal + destVal

	c.SetFlag("N", sum&0x8000 == 0x8000)
	c.SetFlag("Z", sum == 0)
	c.SetFlag("V",
		!((sourceVal^destVal)&0x8000 == 0x8000) && ((destVal^sum)&0x8000 == 0x8000))
	c.SetFlag("C", int(sourceVal)+int(destVal) > 0xffff)
	c.mmunit.WriteMemoryWord(virtAddr, sum)
}

// subtract (16)
func (c *CPU) subOp(instruction uint16) {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(source)
	virtAddr := c.GetVirtualAddress(dest, 0)
	destVal := c.mmunit.ReadMemoryWord(virtAddr) & 0xFFFF

	res := destVal + (^(sourceVal) + 1)
	c.SetFlag("C", sourceVal > destVal)
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", res&0x8000 == 0x8000)
	c.SetFlag("V",
		((sourceVal^destVal)&0x8000 == 0x8000) && !((destVal^res)&0x8000 == 0x8000))
	c.mmunit.WriteMemoryWord(virtAddr, res)
}

// bit (3)
func (c *CPU) bitOp(instruction uint16) {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(source)
	destVal := c.readWord(dest)

	res := sourceVal & destVal
	c.SetFlag("V", false)
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", (res&0x8000) > 0)
}

func (c *CPU) bitbOp(instruction uint16) {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceAddr := c.GetVirtualAddress(source, 1)
	sourceVal := c.mmunit.ReadMemoryByte(sourceAddr)
	destAddr := c.GetVirtualAddress(dest, 1)
	destVal := c.mmunit.ReadMemoryByte(destAddr)

	res := sourceVal & destVal
	c.SetFlag("V", false)
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", (res&0x80) > 0)
}

// bit clear (4)
func (c *CPU) bicOp(instruction uint16) {
	source := (instruction >> 6) & 077
	dest := instruction & 077
	sourceVal := c.readWord(source)
	destAddr := c.GetVirtualAddress(dest, 0)
	destVal := c.mmunit.ReadMemoryWord(destAddr)

	destVal = destVal & (^sourceVal)
	c.SetFlag("V", false)
	c.SetFlag("N", (destVal&0x8000) > 0)
	c.SetFlag("Z", destVal == 0)
	c.mmunit.WriteMemoryWord(destAddr, destVal)
}

func (c *CPU) bicbOp(instruction uint16) {
	source := (instruction >> 6) & 077
	dest := instruction & 077

	sourceVal := c.readByte(source)
	destAddr := c.GetVirtualAddress(dest, 1)
	destVal := c.mmunit.ReadMemoryByte(destAddr)
	destVal = destVal & (^sourceVal)
	c.SetFlag("V", false)
	c.SetFlag("N", (destVal&0x80) == 0x80)
	c.SetFlag("Z", destVal == 0)
	c.mmunit.WriteMemoryByte(destAddr, destVal)
}

// bit inclusive or (5)
func (c *CPU) bisOp(instruction uint16) {
	source := (instruction >> 6) & 077
	dest := instruction & 077

	sourceVal := c.readWord(source)
	virtAddr := c.GetVirtualAddress(dest, 0)
	destVal := c.mmunit.ReadMemoryWord(virtAddr) & 0xFFFF

	destVal = destVal | sourceVal
	c.SetFlag("V", false)
	c.SetFlag("N", (destVal&0x8000) > 0)
	c.SetFlag("Z", destVal == 0)
	c.mmunit.WriteMemoryWord(virtAddr, destVal)
}

func (c *CPU) bisbOp(instruction uint16) {
	source := (instruction >> 6) & 077
	dest := instruction & 077
	sourceAddr := c.GetVirtualAddress(source, 1)
	destAddr := c.GetVirtualAddress(dest, 1)
	sourceVal := c.mmunit.ReadMemoryByte(sourceAddr)
	destVal := c.mmunit.ReadMemoryByte(destAddr)

	destVal = sourceVal | destVal
	c.SetFlag("V", false)
	c.SetFlag("N", (destVal&0x80) == 0x80)
	c.SetFlag("Z", destVal == 0)
	c.mmunit.WriteMemoryByte(destAddr, destVal)
}

// RDD opcodes:

// jsr - jump to subroutine
func (c *CPU) jsrOp(instruction uint16) {
	register := (instruction >> 6) & 7
	destination := instruction & 077
	val := c.GetVirtualAddress(destination, 0)

	// check if destination is register address. it shouldn't be
	if (val & 0177770) == 0170000 {
		panic("JSR to register. That should not happen")
	}

	c.Push(c.Registers[register])
	c.Registers[register] = c.Registers[7]
	c.Registers[7] = val
}

// multiply (070), EIS option
// N and Z flags set as in other implementation
// but as far as I can tell, not as described by Digital
func (c *CPU) mulOp(instruction uint16) {
	sourceOp := instruction & 077
	sourceReg := (instruction >> 6) & 7

	val1 := int(c.Registers[sourceReg])
	if val1&0x8000 == 0x8000 {
		val1 = -((0xFFFF ^ val1) + 1)
	}
	val2 := int(c.readWord(sourceOp))
	if val2&0x8000 == 0x8000 {
		val2 = -((0xFFFF ^ val2) + 1)
	}
	res := int64(val1) * int64(val2)
	c.Registers[sourceReg] = uint16(res >> 16)
	c.Registers[sourceReg|1] = uint16(res & 0xFFFF)
	c.SetFlag("N", res&0xFFFFFFFF == 0)
	c.SetFlag("Z", res&0x80000000 == 0x80000000)
	c.SetFlag("C", (res < (1<<15)) || (res >= (1<<15)-1))
}

// divide (071)
func (c *CPU) divOp(instruction uint16) {
	register := (instruction >> 6) & 7
	source := instruction & 077

	// div operates on a 32 bit digit combined in Register + Register | 1
	val1 := (uint32(c.Registers[register]) << 16) | uint32(c.Registers[register|1])
	val2 := uint32(c.readWord(source))

	if val2 == 0 {
		c.SetFlag("C", true)
		return
	}

	if val1/val2 >= 0x10000 {
		c.SetFlag("V", true)
		return
	}

	c.Registers[register] = uint16((val1 / val2) & 0xFFFF)
	c.Registers[register|1] = uint16((val1 % val2) & 0xFFFF)
	c.SetFlag("Z", c.Registers[register] == 0)
	c.SetFlag("N", c.Registers[register]&0100000 == 0100000)
	c.SetFlag("V", val1 == 0)
}

// shift arithmetically
func (c *CPU) ashOp(instruction uint16) {
	var result uint16

	register := (instruction >> 6) & 7

	// offset is the lower 6 bits of the source operand
	offset := c.readWord(instruction&077) & 077
	source := c.Registers[register]

	// negative number -> shift right
	if (offset & 040) != 0 {
		offset = (077 ^ offset) + 1
		if source&0x8000 == 0x8000 {
			result = 0xFFFF ^ (0xFFFF >> offset)
			result |= source >> offset
		} else {
			result = source >> offset
		}
		shift := uint16(1) << (offset - 1)
		c.SetFlag("C", source&shift == shift)
	} else {
		result = (source << offset) & 0xFFFF
		shift := uint16(1) << (16 - offset)
		c.SetFlag("C", source&shift == shift)
	}

	// V flag set if sign changed:
	c.SetFlag("V", (source&0x8000) != (result&0x8000))
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", result&0x8000 == 0x8000)
	c.Registers[register] = result
}

// arithmetic shift combined (EIS option)
func (c *CPU) ashcOp(instruction uint16) {

	var result uint32
	destAddr := c.GetVirtualAddress(instruction&077, 0)
	offset := uint8(c.mmunit.ReadMemoryWord(destAddr) & 077)
	if offset == 0 {
		return
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
}

// xor
func (c *CPU) xorOp(instruction uint16) {
	sourceVal := c.Registers[(instruction>>6)&7]
	destAddr := c.GetVirtualAddress(instruction&077, 0)
	destVal := c.mmunit.ReadMemoryWord(destAddr)

	res := sourceVal ^ destVal

	c.SetFlag("N", res < 0)
	c.SetFlag("Z", res == 0)
	c.SetFlag("V", false)
	c.mmunit.WriteMemoryWord(destAddr, res)
}

// sob - subtract one and branch (if not equal 0)
// if value of the register sourceReg is not 0, subtract
// twice the value of the offset (lowest 6 bits) from the SP
func (c *CPU) sobOp(instruction uint16) {
	sourceReg := (instruction >> 6) & 7
	c.Registers[sourceReg] = (c.Registers[sourceReg] - 1) & 0xffff
	if c.Registers[sourceReg] != 0 {
		c.Registers[7] = (c.Registers[7] - ((instruction & 077) << 1)) & 0xffff
	}
}

// trap opcodes:
// todo: something fishy is going on here
// todo: add test
func (c *CPU) trapOpcode(vector uint16) {
	prevPs := c.unibus.Psw.Get()
	c.SwitchMode(psw.KernelMode)

	// push current PS and PC to stack
	c.Push(prevPs)
	c.Push(c.Registers[7])

	// load PC and PS from trap vector location
	c.Registers[7] = c.mmunit.ReadMemoryWord(vector)
	newPsw := c.mmunit.ReadMemoryWord(vector + 2)
	if prevPs&(1<<14) > 0 {
		newPsw |= (1 << 13) | (1 << 12)
	}

	// todo -> can the new PSW set the cpu to the user mode?

	c.unibus.Psw.Set(newPsw)
}

// emt - emulator trap - trap vector hardcoded to location 32
func (c *CPU) emtOp(_ uint16) {
	c.trapOpcode(030)
}

// trap vector for TRAP is hardcoded for all PDP11s to memory location 034
func (c *CPU) trapOp(_ uint16) {
	c.trapOpcode(034)
}

// Single Register opcodes
// rts - return from subroutine
func (c *CPU) rtsOp(instruction uint16) {
	register := instruction & 7

	// load Program Counter from register passed in instruction
	c.Registers[7] = c.Registers[register]

	// load word popped from processor stack to "register"
	c.Registers[register] = c.Pop()
}

// clear flag opcodes
// covers following operations: CLN, CLZ, CLV, CLC, CCC
func (c *CPU) clearFlagOp(instruction uint16) {

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
}

// set flag opcodes
// covers following operations: SEN, SEZ, SEV, SEC, SCC
func (c *CPU) setFlagOp(instruction uint16) {
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
}
