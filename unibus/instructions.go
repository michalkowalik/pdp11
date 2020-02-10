package unibus

import (
	"pdp/interrupts"
)

// Definition of all PDP-11 CPU instructions
// All should follow the func (*CPU) (int16) signature

// single operand cpu instructions:
// clr Word / Byte
func (c *CPU) clrOp(instruction uint16) {
	if instruction&0100000 == 0100000 {
		dstAddr, _ := c.GetVirtualByMode(instruction&077, 1)
		c.mmunit.WriteMemoryByte(dstAddr, 0)
	} else {
		dstAddr, _ := c.GetVirtualByMode(instruction&077, 0)
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
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 0)
	dst := c.mmunit.ReadMemoryWord(dstAddr)

	c.SetFlag("N", dst&0x8000 == 0x8000)
	c.SetFlag("Z", dst == 0)
	c.SetFlag("C", true)
	c.SetFlag("V", false)
	c.mmunit.WriteMemoryWord(dstAddr, ^dst)
}

func (c *CPU) combOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 0)
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
	virtAddr, _ := c.GetVirtualByMode(dest, 0)
	val := (c.mmunit.ReadMemoryWord(virtAddr) + 1) & 0xFFFF
	c.mmunit.WriteMemoryWord(virtAddr, val)

	c.SetFlag("Z", val == 0)
	c.SetFlag("N", val&0x8000 == 0x8000)
}

func (c *CPU) incbOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 1)
	val := c.mmunit.ReadMemoryByte(dstAddr)
	res := (val + 1) & 0xFF
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", res&0x80 == 0x80)
	c.SetFlag("V", val == 0x7F)
	c.mmunit.WriteMemoryByte(dstAddr, val)
}

// dec - decrement dst
// TODO: it should look like INC
func (c *CPU) decOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 0)
	val := c.mmunit.ReadMemoryWord(dstAddr)
	res := (val - 1) & 0xFFFF

	c.SetFlag("Z", res == 0)
	c.SetFlag("N", res&0x8000 == 0x8000)
	c.SetFlag("V", val == 0100000)
	c.mmunit.WriteMemoryWord(dstAddr, res)
}

func (c *CPU) decbOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 1)
	val := c.mmunit.ReadMemoryByte(dstAddr)
	res := (val - 1) & 0xFF

	c.SetFlag("Z", res == 0)
	c.SetFlag("N", res&0x80 == 0x80)
	c.SetFlag("V", val == 0x80)
	c.mmunit.WriteMemoryByte(dstAddr, res)
}

// neg - negate dst
// replace the contents of the destination address
// by it's 2 complement. 01000000 is replaced by itself
func (c *CPU) negOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(dstAddr)
	result := ^dest + 1
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", int16(result) < 0)
	c.SetFlag("V", result == 0x8000)
	c.SetFlag("C", result != 0)
	c.mmunit.WriteMemoryWord(dstAddr, result)
}

func (c *CPU) negbOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 1)
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
		dstAddr, _ = c.GetVirtualByMode(instruction&077, 1)
		dst = uint16(c.mmunit.ReadMemoryByte(dstAddr))
		msb = 0x80
		ov = 0177
		oc = 0xFF
	} else {
		dstAddr, _ = c.GetVirtualByMode(instruction&077, 0)
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

// sbc - substract carry
func (c *CPU) sbcOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(dstAddr)
	result := dest
	if c.GetFlag("C") {
		result = result - 1
		c.mmunit.WriteMemoryWord(dstAddr, result)
	}

	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("Z", result == 0)
	c.SetFlag("V", dest == 0x8000)
	c.SetFlag("C", (dest != 0) || c.GetFlag("C"))
}

func (c *CPU) sbcbOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(dstAddr)
	result := dest
	if c.GetFlag("C") {
		result = result - 1
		c.mmunit.WriteMemoryByte(dstAddr, result)
	}

	c.SetFlag("N", (result&0x80) == 0x80)
	c.SetFlag("Z", result == 0)
	c.SetFlag("V", dest == 0x80)
	c.SetFlag("C", (dest != 0) || c.GetFlag("C"))
}

// tst - sets the condition codes N and Z according to the contents
// of the destination address
func (c *CPU) tstOp(instruction uint16) {
	dest := c.readWord(uint16(instruction & 077))
	c.SetFlag("Z", dest == 0)
	c.SetFlag("N", (dest&0x8000) > 0)
	c.SetFlag("V", false)
	c.SetFlag("C", false)
}

func (c *CPU) tstbOp(instruction uint16) {
	dest := c.readByte(uint16(instruction & 077))

	c.SetFlag("Z", dest == 0)
	c.SetFlag("N", (dest&0x80) > 0)
	c.SetFlag("V", false)
	c.SetFlag("C", false)
}

// asr - arithmetic shift right
// 	Shifts all bits of the destination right one place. Bit 15
// is replicated. The C-bit is loaded from bit 0 of the destination.
// ASR performs signed division of the destination by two.
func (c *CPU) asrOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(dstAddr)
	result := (dest & 0x8000) | (dest >> 1)
	c.mmunit.WriteMemoryWord(dstAddr, result)

	c.SetFlag("C", (dest&1) == 1)
	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("Z", result == 0)

	// V flag is a XOR on C and N flag, but golang doesn't provide boolean XOR
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
}

func (c *CPU) asrbOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(dstAddr)
	result := (dest & 0x80) | (dest >> 1)
	c.mmunit.WriteMemoryByte(dstAddr, result)

	c.SetFlag("C", (dest&1) == 1)
	c.SetFlag("N", (result&0x80) == 0x80)
	c.SetFlag("Z", result == 0)

	// V flag is a XOR on C and N flag, but golang doesn't provide boolean XOR
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
}

// asl - arithmetic shift left
// Shifts all bits of the destination left one place. Bit 0 is
// loaded with an 0. The C·bit of the status word is loaded from
// the most significant bit of the destination. ASL performs a
// signed multiplication of the destination by 2 with overflow indication.
func (c *CPU) aslOp(instruction uint16) {
	destAddr, _ := c.GetVirtualByMode(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(destAddr)
	result := dest << 1
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("C", (dest&0x8000) == 0x8000)
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
	c.mmunit.WriteMemoryWord(destAddr, result)
}

func (c *CPU) aslbOp(instruction uint16) {
	destAddr, _ := c.GetVirtualByMode(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(destAddr)
	result := dest << 1
	c.SetFlag("Z", result == 0)
	c.SetFlag("N", (result&0x80) == 0x80)
	c.SetFlag("C", (dest&0x80) == 0x80)
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
	c.mmunit.WriteMemoryByte(destAddr, result)
}

// ror - rotate right
// Rotates all bits of the destination right one place. Bit 0 is
// loaded into the C-bit and the previous contents of the C-bit
// are loaded into bit 15 of the destination.
func (c *CPU) rorOp(instruction uint16) {
	destAddr, _ := c.GetVirtualByMode(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(destAddr)
	cBit := (dest & 1) << 15
	result := (dest >> 1) | cBit
	c.SetFlag("N", (result&0x8000) == 0x8000)
	c.SetFlag("Z", result == 0)
	c.SetFlag("C", cBit > 0)
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
	c.mmunit.WriteMemoryWord(destAddr, result)
}

func (c *CPU) rorbOp(instruction uint16) {
	destAddr, _ := c.GetVirtualByMode(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(destAddr)
	cBit := (dest & 1) << 7
	result := (dest >> 1) | cBit
	c.SetFlag("N", (result&0x80) == 0x80)
	c.SetFlag("Z", result == 0)
	c.SetFlag("C", cBit > 0)
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
	c.mmunit.WriteMemoryByte(destAddr, result)
}

// rol - rorare left
// : Rotate all bits of the destination left one place. Bit 15
// is loaded into the C·bit of the status word and the previous
// contents of the C-bit are loaded into Bit 0 of the destination.
func (c *CPU) rolOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 0)
	dest := c.mmunit.ReadMemoryWord(dstAddr)
	res := dest << 1

	if c.GetFlag("C") {
		res |= 1
	}
	c.SetFlag("C", (dest&0x8000) == 0x8000)
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", (res&0x8000) == 0x8000)
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
	c.mmunit.WriteMemoryWord(dstAddr, res)
}

func (c *CPU) rolbOp(instruction uint16) {
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 1)
	dest := c.mmunit.ReadMemoryByte(dstAddr)
	res := dest << 1

	if c.GetFlag("C") {
		res |= 1
	}
	c.SetFlag("C", (dest&0x80) == 0x80)
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", (res&0x80) == 0x80)
	c.SetFlag("V", (c.GetFlag("C") != c.GetFlag("N")) == true)
	c.mmunit.WriteMemoryByte(dstAddr, res)
}

// jmp - jump to address:
func (c *CPU) jmpOp(instruction uint16) {
	dest, _ := c.GetVirtualByMode(instruction&077, 0) // c.readWord(uint16(instruction & 077))
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
	dstAddr, _ := c.GetVirtualByMode(instruction&077, 0)
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
	c.Registers[6] = c.Registers[7] + uint16(instruction&0xFFFF)<<1
	c.Registers[7] = c.Registers[5]
	c.Registers[5] = c.Pop()
}

// mfpi - move from previous instruction space
func (c *CPU) mfpiOp(instruction uint16) {

	var val uint16
	dest, err := c.GetVirtualByMode(instruction&077, 0)
	if err != nil {
		panic("MFPI: could not resolve virtual address")
	}

	curUser := c.mmunit.Psw.GetMode()
	prevUser := c.mmunit.Psw.GetPreviousMode()

	switch {
	case dest == 0170006:
		if curUser == prevUser {
			val = c.Registers[6]
		} else {
			if curUser == 0 { // kernel
				val = c.UserStackPointer
			} else {
				val = c.KernelStackPointer
			}
		}
	// if register:
	case dest&0177770 == 0170000:
		panic("MFPI attended on Register address")
	default:
		physicalAddress := c.mmunit.mapVirtualToPhysical(dest, false, prevUser)
		val =
			c.mmunit.ReadWordByPhysicalAddress(physicalAddress)
	}

	c.Push(val)
	c.mmunit.Psw.Set(c.mmunit.Psw.Get() & 0xFFF0)
	c.SetFlag("C", true)
	c.SetFlag("N", val&0x8000 == 0x8000)
	c.SetFlag("Z", val == 0)
}

// mtpi - move to previous instruction space
func (c *CPU) mtpiOp(instruction uint16) {
	dest, err := c.GetVirtualByMode(instruction&077, 0)
	if err != nil {
		panic("INC: Can't obtain virtual address")
	}
	val := c.Pop()

	curUser := c.mmunit.Psw.GetMode()
	prevUser := c.mmunit.Psw.GetPreviousMode()

	switch {
	case dest == 0170006:
		if curUser == prevUser {
			c.Registers[6] = val
		} else {
			if curUser == 0 {
				c.UserStackPointer = val
			} else {
				c.KernelStackPointer = val
			}
		}
	case dest&0177770 == 0170000:
		panic("MTPI attended on Register address")
	default:
		sourceAddress := c.mmunit.mapVirtualToPhysical(dest, false, prevUser)
		c.mmunit.WriteWordByPhysicalAddress(sourceAddress, val)
	}

	c.mmunit.Psw.Set(c.mmunit.Psw.Get() & 0xFFFF)
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

	if err := c.writeWord(uint16(instruction&077), uint16(res)); err != nil {
		panic(err)
	}

	c.SetFlag("Z", !c.GetFlag("N"))
}

// double operand cpu instructions:

// move (1)
// There's a subtle problem: what if MOV is used to set the PSW?
// because it will set the flags according to the value of the operand,
// possibly rewriting the flags in the process.
func (c *CPU) movOp(instruction uint16) {

	source := (instruction & 07700) >> 6
	dest := instruction & 077
	sourceVal := c.readWord(uint16(source))

	c.SetFlag("N", (sourceVal&0100000) > 0)
	c.SetFlag("Z", sourceVal == 0)
	// V is always cleared by MOV
	c.SetFlag("V", false)
	c.writeWord(uint16(dest), sourceVal)
}

// movb
// TODO: Finish implementation
func (c *CPU) movbOp(instruction uint16) {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceAddr, _ := c.GetVirtualByMode(source, 1)
	sourceVal := c.mmunit.ReadMemoryByte(sourceAddr)
	destAddr, _ := c.GetVirtualByMode(dest, 1)
	c.SetFlag("Z", sourceVal == 0)
	c.SetFlag("V", false)
	c.SetFlag("N", (sourceVal&0x80) > 0)

	c.mmunit.WriteMemoryByte(destAddr, sourceVal)
}

// misc instructions (decode all bits)
// halt
func (c *CPU) haltOp(instruction uint16) {
	c.State = HALT
}

// bpt - breakpoint trap
func (c *CPU) bptOp(instruction uint16) {
	// 14 is breakpoint trap vector
	c.Trap(interrupts.Trap{Vector: 014, Msg: "Breakpoint"})
}

// iot - i/o trap
func (c *CPU) iotOp(instruction uint16) {
	c.Trap(interrupts.Trap{Vector: 020, Msg: "i/o"})
}

// rti - return from interrupt
func (c *CPU) rtiOp(instruction uint16) {
	c.Registers[7] = c.Pop()
	val := c.Pop()
	if c.mmunit.Psw.GetMode() == UserMode {
		val &= 047
		val |= c.mmunit.Psw.Get() & 0177730
	}
	c.mmunit.Psw.Set(val)
}

// rtt - return from trap
func (c *CPU) rttOp(instruction uint16) {
	c.rtiOp(instruction)
}

// wait for interrupt
// check for interrupts here!!
func (c *CPU) waitOp(instruction uint16) {
	c.State = WAIT
}

// Sends INIT on UNIBUS for 10ms. All devices on the UNIBUS are reset and power up
// TODO: user mode?
func (c *CPU) resetOp(instruction uint16) {
	c.mmunit.unibus.Rk01.Reset()
	c.mmunit.unibus.TermEmulator.ClearTerminal()
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
		sourceVal = uint16(c.readByte(uint16(source)))
		destVal = uint16(c.readByte(uint16(dest)))
	} else {
		sourceVal = c.readWord(uint16(source))
		destVal = c.readWord(uint16(dest))
	}
	res := sourceVal + (^(destVal) + 1)

	c.SetFlag("N", (res&msb) > 0)
	c.SetFlag("Z", res == 0)
	c.SetFlag("C", sourceVal < destVal)
	c.SetFlag("V", (sourceVal^destVal)&msb == msb && !((destVal^res)&msb == msb))
}

//add (6)
func (c *CPU) addOp(instruction uint16) {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	virtAddr, _ := c.GetVirtualByMode(dest, 0)
	destVal := c.mmunit.ReadMemoryWord(virtAddr)
	sum := sourceVal + destVal

	c.SetFlag("N", sum&0x8000 == 0x8000)
	c.SetFlag("Z", sum == 0)
	c.SetFlag("V",
		!((sourceVal^destVal)&0x8000 == 0x8000) && ((destVal^sum)&0x8000 == 0x8000))
	c.SetFlag("C", int(sourceVal)+int(destVal) > 0xffff)
	c.mmunit.WriteMemoryWord(virtAddr, sum)
}

// substract (16)
func (c *CPU) subOp(instruction uint16) {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	virtAddr, err := c.GetVirtualByMode(dest, 0)
	if err != nil {
		panic("INC: Can't obtain virtual address")
	}
	destVal := c.mmunit.ReadMemoryWord(virtAddr) & 0xFFFF

	res := destVal + (^(sourceVal) + 1)
	c.SetFlag("C", sourceVal > destVal)
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", res&0x8000 == 0x8000)
	c.SetFlag("V",
		((sourceVal^destVal)&0x8000 == 0x8000) && !((destVal^res)&0x8000 == 0x8000))
	c.mmunit.WriteMemoryWord(virtAddr, res)
}

//bit (3)
func (c *CPU) bitOp(instruction uint16) {
	source := (instruction & 07700) >> 6
	dest := instruction & 077

	sourceVal := c.readWord(uint16(source))
	destVal := c.readWord(uint16(dest))

	res := sourceVal & destVal
	c.SetFlag("V", false)
	c.SetFlag("Z", res == 0)
	c.SetFlag("N", (res&0x8000) > 0)
}

func (c *CPU) bitbOp(instruction uint16) {
	source := (instruction & 07700) >> 6
	dest := instruction & 7

	sourceVal := c.readByte(source)
	destAddr, _ := c.GetVirtualByMode(dest, 1)
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
	sourceVal := c.readWord(uint16(source))
	destAddr, _ := c.GetVirtualByMode(dest, 0)
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
	destAddr, _ := c.GetVirtualByMode(dest, 1)
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

	sourceVal := c.readWord(uint16(source))
	virtAddr, _ := c.GetVirtualByMode(dest, 0)
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
	sourceAddr, _ := c.GetVirtualByMode(source, 1)
	destAddr, _ := c.GetVirtualByMode(dest, 1)
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
	destination := uint16(instruction & 077)
	val, _ := c.GetVirtualByMode(destination, 0)

	c.Push(uint16(c.Registers[register]))
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
	source := uint16(instruction & 077)

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

	// offset is the lower 6 bits of the source
	offset := c.readWord(uint16(instruction&077)) & 077
	source := uint16(c.Registers[register])

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
	offset := uint16(instruction & 077)
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
}

// xor
func (c *CPU) xorOp(instruction uint16) {
	sourceVal := c.Registers[(instruction>>6)&7]
	destAddr, _ := c.GetVirtualByMode(instruction&077, 0)
	destVal := c.mmunit.ReadMemoryWord(destAddr)

	res := sourceVal ^ destVal

	c.SetFlag("N", res < 0)
	c.SetFlag("Z", res == 0)
	c.SetFlag("V", false)
	c.mmunit.WriteMemoryWord(destAddr, res)
}

// sob - substract one and branch (if not equal 0)
// if value of the register sourceReg is not 0, susbtract
// twice the value of the offset (lowest 6 bits) from the SP
func (c *CPU) sobOp(instruction uint16) {
	sourceReg := (instruction >> 6) & 7
	c.Registers[sourceReg] = (c.Registers[sourceReg] - 1) & 0xffff
	if c.Registers[sourceReg] != 0 {
		c.Registers[7] = (c.Registers[7] - ((uint16(instruction) & 077) << 1)) & 0xffff
	}
}

// trap opcodes:
// emt - emulator trap - trap vector hardcoded to location 32
func (c *CPU) emtOp(instruction uint16) {
	c.Trap(interrupts.Trap{Vector: 32, Msg: "emt"})
}

// trap
// trap vector for TRAP is hardcoded for all PDP11s to memory location 34
func (c *CPU) trapOp(instruction uint16) {
	c.Trap(interrupts.Trap{Vector: 34, Msg: "TRAP"})
}

// Single Register opcodes
// rts - return from subroutine
func (c *CPU) rtsOp(instruction uint16) {
	register := instruction & 7

	// load Program Counter from register passed in instruction
	c.Registers[7] = c.Registers[register]

	// load word popped from processor stack to "register"
	val := c.Pop()
	c.Registers[register] = val
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
