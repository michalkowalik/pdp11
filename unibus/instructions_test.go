package unibus

import (
	"errors"
	"fmt"
	"os"
	"pdp/console"
	"pdp/psw"
	"testing"
)

// flags is a struct used to assert cpu flags settings
type flags struct {
	c bool
	v bool
	z bool
	n bool
}

// global shared resources: CPU, memory etc.
var c *CPU
var u *Unibus
var memory [0x400000]byte // 64KB of memory is all everyone needs

// TestMain to resucure -> initialize memory and CPU
func TestMain(m *testing.M) {
	mmu := &MMU18Bit{}
	p := psw.PSW(0)
	mmu.Psw = &p
	c = NewCPU(mmu)

	var cons console.Console
	cons = console.NewSimple()
	u = New(&p, nil, &cons)

	os.Exit(m.Run())
}

func TestCPU_clrOp(t *testing.T) {
	type args struct {
		instruction uint16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"value in register", args{05000}, false},
		{"address in register", args{05011}, false},
	}
	u.PdpCPU.Registers[0] = 0xff
	u.PdpCPU.Registers[1] = 0xfe

	// and let's give CPU stack some place to breath:
	u.PdpCPU.Registers[6] = 0xfe

	// come back here!!
	u.Mmu.Memory[0xfe] = 2

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u.PdpCPU.clrOp(tt.args.instruction)
			// also: check if value is really 0:
			op := uint16(tt.args.instruction) & 077
			t.Logf("instruction: %x, op: %x\n", tt.args.instruction, op)
			w := u.PdpCPU.readWord(op)
			if w != 0 {
				t.Errorf("CPU.clrOp() -> destination for %v is set to %x", op, w)
			}
		})
	}
}

func TestCPU_addOp(t *testing.T) {
	type args struct {
		instruction uint16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantRes int16
	}{
		{"dst in register, src memory value", args{061100}, false, 0x1fe},
	}

	u.PdpCPU.Registers[0] = 0xff
	u.PdpCPU.Registers[1] = 0xfe
	u.PdpCPU.Registers[2] = 0
	u.PdpCPU.Registers[3] = 2
	u.PdpCPU.mmunit.Memory[0x7f] = 0xff
	u.PdpCPU.mmunit.Memory[0] = 2
	u.PdpCPU.mmunit.Memory[2] = 0x300

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u.PdpCPU.addOp(tt.args.instruction)

			//check value
			w := u.PdpCPU.readWord(uint16(tt.args.instruction & 077))
			t.Logf("Value at dst: %x\n", w)
			if int16(w) != tt.wantRes {
				t.Errorf("expected %x, got %x", tt.wantRes, w)
			}
		})
	}
}

func TestCPU_movOp(t *testing.T) {
	type args struct {
		instruction uint16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		dst     int16
	}{
		{"move from memory to register", args{011001}, false, 4},
	}

	c.Registers[0] = 0xfe
	c.Registers[1] = 0
	c.mmunit.Memory[0x7f] = uint16(4)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.movOp(tt.args.instruction)
			d := c.readWord(uint16(tt.args.instruction & 077))

			if int16(d) != tt.dst {
				t.Logf("destination addr: %x\n", tt.args.instruction&077)
				t.Errorf("Expected destination: %x, got %x\n", tt.dst, d)
			}
		})
	}
}

func TestCPU_movbOp(t *testing.T) {
	type args struct {
		instruction uint16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		dst     uint16
	}{
		{"MOV move from memory to register", args{0111001}, false, 0},
	}

	u.PdpCPU.Registers[0] = 0xfe
	u.PdpCPU.Registers[1] = 0
	u.Mmu.Memory[0x7f] = uint16(4)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opcode := u.PdpCPU.Decode(tt.args.instruction)
			opcode(tt.args.instruction)
			d := u.PdpCPU.readWord(tt.args.instruction & 077)

			if d != tt.dst {
				t.Logf("destination addr: %x\n", tt.args.instruction&077)
				t.Errorf("Expected destination: %x, got %x\n", tt.dst, d)
			}
		})
	}
}

// TODO: finish test implementation
// tests:
// - offset 0 -> no res
// - negative offset
// - positive offset
// - with and without the V flag set?
// - odd register number -> basically a rotate
func TestCPU_comOp(t *testing.T) {
	type args struct {
		instruction uint16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		dst     uint16
	}{
		{"complement dst on value in register", args{005100}, false, 0xff0f},
	}

	u.PdpCPU.Registers[0] = 0xf0
	u.Mmu.Memory[0xff] = uint16(4)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opcode := u.PdpCPU.Decode(uint16(tt.args.instruction))
			opcode(tt.args.instruction)
			d := u.PdpCPU.readWord(uint16(tt.args.instruction & 077))

			if d != tt.dst {
				t.Logf("destination addr: %x\n", tt.args.instruction&077)
				t.Errorf("Expected destination: %x, got %x\n", tt.dst, d)
			}
		})
	}
}

func TestCPU_incOp(t *testing.T) {
	type args struct {
		instruction uint16
	}
	tests := []struct {
		name    string
		args    args
		regVal  uint16
		dst     uint16
		wantErr bool
		vFlag   bool
		zFlag   bool
		nFlag   bool
	}{

		{"INC on 0x7FFF should set V and N flag",
			args{05200}, 0x7FFF, 0x8000, false, false, false, true},
		{"INC on 0x0000 should set no flags",
			args{05200}, 0, 1, false, false, false, false},
		{"INC on 0xffff should set Z flag",
			args{05200}, 0xffff, 0, false, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u.PdpCPU.Registers[0] = tt.regVal
			instruction := u.PdpCPU.Decode(uint16(tt.args.instruction))
			instruction(tt.args.instruction)

			d := u.PdpCPU.readWord(uint16(tt.args.instruction & 077))
			if d != tt.dst {
				t.Errorf("Expected value: %x, got: %x\n", tt.dst, d)
			}
			if v := c.GetFlag("V"); v != tt.vFlag {
				t.Errorf("Overflow flag error. Expected %v, got %v\n", tt.vFlag, v)
			}
			if z := c.GetFlag("Z"); z != tt.zFlag {
				t.Errorf("Zero flag error. expected %v, got %v\n", tt.zFlag, z)
			}
			if n := c.GetFlag("N"); n != tt.nFlag {
				t.Errorf("Negative flag error. Expected %v, got %v\n", tt.nFlag, n)
			}
		})
	}
}

func TestCPU_negOp(t *testing.T) {
	type args struct {
		instruction uint16
	}
	tests := []struct {
		name    string
		args    args
		regVal  uint16
		dst     uint16
		wantErr bool
		vFlag   bool
		zFlag   bool
		nFlag   bool
		cFlag   bool
	}{
		{"Neg on -1 should give 1 and set C flag",
			args{05400}, 0xffff, 1, false, false, false, false, true},
		{"Neg on 10 should give  -10 and set C and N flags",
			args{05400}, 0xA, 0xfff6, false, false, false, true, true},
		{"Neg on 0 should give 0, set Z and unset C",
			args{05400}, 0, 0, false, false, true, false, false},
		{"Neg on 0x8000 should cause overflow and set N flag",
			args{05400}, 0x8000, 0x8000, false, true, false, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Registers[0] = tt.regVal
			instruction := c.Decode(uint16(tt.args.instruction))
			instruction(tt.args.instruction)
			if c.Registers[0] != tt.dst {
				t.Errorf("\"%s\" ERROR: expected %v, got %v\n",
					tt.name, tt.dst, c.Registers[0])
			}
			if z := c.GetFlag("Z"); z != tt.zFlag {
				t.Errorf("Z flag error. Expected %v, got %v\n", tt.zFlag, z)
			}
			if c := c.GetFlag("C"); c != tt.cFlag {
				t.Errorf("C flag error. Expected %v, got %v\n", tt.cFlag, c)
			}
			if n := c.GetFlag("N"); n != tt.nFlag {
				t.Errorf("N flag error. Expected %v, got %v\n", tt.nFlag, n)
			}
			if v := c.GetFlag("V"); v != tt.vFlag {
				t.Errorf("V flag error. Expected %v, got %v\n", tt.vFlag, v)
			}
		})
	}
}

func TestCPU_adcOp(t *testing.T) {
	type args struct {
		instruction uint16
	}
	tests := []struct {
		name      string
		args      args
		regVal    uint16
		dst       uint16
		origCFlag bool
		wantErr   bool
		vFlag     bool
		zFlag     bool
		nFlag     bool
		cFlag     bool
	}{
		{"ADC on 0xFFFF with C set should return 0 and set Z and C flags",
			args{05500}, 0xffff, 0, true, false, false, true, false, true},
		{"ADC on 0x7FFF with C set should return 0x8000 and set V flag",
			args{05500}, 0x7fff, 0x8000, true, false, true, false, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Registers[0] = tt.regVal
			c.SetFlag("C", tt.origCFlag)
			instruction := c.Decode(uint16(tt.args.instruction))
			instruction(tt.args.instruction)
			if c.Registers[0] != tt.dst {
				t.Errorf("ADC returned unexpected result. expected %v, got %v\n",
					tt.dst, c.Registers[0])
			}
			if z := c.GetFlag("Z"); z != tt.zFlag {
				t.Errorf("Z flag error. Expected %v, got %v\n", tt.zFlag, z)
			}
			if c := c.GetFlag("C"); c != tt.cFlag {
				t.Errorf("C flag error. Expected %v, got %v\n", tt.cFlag, c)
			}
			if n := c.GetFlag("N"); n != tt.nFlag {
				t.Errorf("N flag error. Expected %v, got %v\n", tt.nFlag, n)
			}
			if v := c.GetFlag("V"); v != tt.vFlag {
				t.Errorf("V flag error. Expected %v, got %v\n", tt.vFlag, v)
			}
		})
	}
}

func TestCPU_xorOp(t *testing.T) {
	type args struct {
		instruction uint16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantRes uint16
	}{
		{"dst value in REG", args{074002}, false, 000325},
	}

	u.PdpCPU.Registers[0] = 001234
	u.PdpCPU.Registers[2] = 001111

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u.PdpCPU.xorOp(tt.args.instruction)
			w := u.PdpCPU.readWord(uint16(tt.args.instruction & 077))
			t.Logf("Value at dst: %x \n", w)
			if w != tt.wantRes {
				t.Errorf("expected %x, got %x\n", tt.wantRes, w)
			}
		})
	}
}

func TestCPU_ashcOp(t *testing.T) {
	type args struct {
		instruction uint16
	}
	tests := []struct {
		name             string
		args             args   // arg value
		rValue           uint16 // selected register value
		rPlusValue       uint16 // r+1 value
		rExpectedVal     uint16
		rPlusExpectedVal uint16
		carrySet         bool
		wantErr          bool
	}{
		{"Even register number, no carry", args{073001}, 1, 1, 2, 2, false, false},
		{"odd  register number, no carry", args{073103}, 1, 1, 8, 8, false, false},
		{"Even register number, carry set", args{073001}, 0xffff, 0xffff, 0xffff, 0xfffe, true, false},
		{"Even register number, no carry, right shift", args{073077}, 2, 2, 1, 1, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetFlag("C", false)
			ashcLoadRegisters(tt.args.instruction, tt.rValue, tt.rPlusValue)
			c.ashcOp(tt.args.instruction)

			// assert the register and flag values after the op
			if err := assertRegistersShifted(
				tt.args.instruction, tt.rExpectedVal, tt.rPlusExpectedVal); err != nil {
				t.Errorf("Register values after shift does not match: %s\n", err)
			}

			// assert carry flag set as expected:
			if c.GetFlag("C") != tt.carrySet {
				t.Errorf("Carry flag false value\n")
			}
		})
	}
}

// helper functions:

// ashcLoadRegisters loads the register values to the register number
// extracted from the 8-6 bits of the opcode
func ashcLoadRegisters(op uint16, rValue, rPlusValue uint16) {
	register := (op >> 6) & 7
	c.Registers[register] = rValue
	c.Registers[register|1] = rPlusValue
}

// assertRegistersShifted checks if the register values ar shifted after
// the ashc operation
// TODO: add right shift - but for now check the left at least
func assertRegistersShifted(op uint16, rValue, rPlusValue uint16) error {
	register := (op >> 6) & 7
	plusRegister := register | 1

	regValue := c.Registers[register]
	regPlusValue := c.Registers[plusRegister]

	// fmt.Printf("regVal: %x, regPlusVal: %x\n", regValue, regPlusValue)

	if regValue != rValue {
		return fmt.Errorf("regValue != shifted rValue : %v vs %v",
			regValue, rValue)
	}

	if regPlusValue != rPlusValue {
		return fmt.Errorf("regPlusValue != shifted rValue: %v vs %v",
			regPlusValue, rPlusValue)
	}

	return nil
}

func TestCPU_ashOp(t *testing.T) {
	type args struct {
		instruction uint16
	}
	tests := []struct {
		name         string
		args         args   // arg value
		rValue       uint16 // selected register value
		rExpectedVal uint16
		carrySet     bool
		wantErr      bool
	}{
		{"left shift, no carry", args{072001}, 1, 2, false, false},
		{"right shift, no carry", args{072077}, 2, 2, true, false},
		{"left shift, carry", args{072001}, 0x8000, 0, true, false},
		{"right shift, carry", args{072077}, 1, 1, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetFlag("C", false)
			register := (tt.args.instruction >> 6) & 7
			c.Registers[register] = tt.rValue
			c.ashOp(tt.args.instruction)

			// assert values of shifted register:
			if c.Registers[register] != tt.rExpectedVal {
				t.Errorf(
					"CPU.ashOp() expected Register value = %v, got %v",
					tt.rExpectedVal,
					c.Registers[register])
			}
			// assert carry flag set
			if c.GetFlag("C") != tt.carrySet {
				t.Errorf("CPU.ashOp() carry flag = %v, expected %v", c.GetFlag("C"), tt.carrySet)
			}
		})
	}
}

// for simplicity sake values are kept in registers directly,
// src is always in R0
// dst in R1
// validity of decoding instructions and fetching from memory is tested in the cpu module
// hence, it's always the same instruction.
func TestCPU_subOp(t *testing.T) {
	// substract: R1 = R1 - R0
	var instruction uint16
	instruction = 0160001

	tests := []struct {
		name    string
		r0Val   int16
		r1Val   int16
		res     int16
		flags   flags
		wantErr bool
	}{
		{"No flags set", 011111, 012345, 01234, flags{false, false, false, false}, false},
	}
	for _, tt := range tests {
		u.PdpCPU.Registers[0] = uint16(tt.r0Val)
		u.PdpCPU.Registers[1] = uint16(tt.r1Val)
		t.Run(tt.name, func(t *testing.T) {
			u.PdpCPU.subOp(instruction)

			// assert value
			if u.PdpCPU.Registers[1] != uint16(tt.res) {
				t.Errorf("CPU.subOp result = %x, expected %x", c.Registers[1], tt.res)
			}

			// check flags
			if err := assertFlags(tt.flags, u.PdpCPU); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func TestCPU_bicOp(t *testing.T) {
	// BIC R0, R1
	var instruction uint16 = 040001

	tests := []struct {
		name    string
		r0Val   uint16
		r1Val   uint16
		res     uint16
		flags   flags
		wantErr bool
	}{
		{"Clear all bits", 0xffff, 0xffff, 0, flags{false, false, true, false}, false},
		{"N flag set", 0x7fff, 0xffff, 0x8000, flags{false, false, false, true}, false},
	}
	for _, tt := range tests {
		u.PdpCPU.Registers[0] = tt.r0Val
		u.PdpCPU.Registers[1] = tt.r1Val
		t.Run(tt.name, func(t *testing.T) {
			u.PdpCPU.bicOp(instruction)

			// assert value
			if u.PdpCPU.Registers[1] != tt.res {
				t.Errorf("CPU.bicOp() r1 = %x, r0 = %x, exp -> %x",
					u.PdpCPU.Registers[1], u.PdpCPU.Registers[0], tt.res)
			}

			// check flags
			if err := assertFlags(tt.flags, u.PdpCPU); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

// helper functions
func assertFlags(flags flags, c *CPU) error {
	passed := true
	errorMsg := ""
	if c.GetFlag("C") != flags.c {
		passed = false
		errorMsg += fmt.Sprintf(" C -> exp %v, got %v ", flags.c, c.GetFlag("C"))
	}
	if c.GetFlag("Z") != flags.z {
		passed = false
		errorMsg += fmt.Sprintf(" Z -> exp %v, got %v ", flags.z, c.GetFlag("Z"))
	}
	if c.GetFlag("N") != flags.n {
		passed = false
		errorMsg += fmt.Sprintf(" N -> exp %v, got %v ", flags.n, c.GetFlag("N"))
	}
	if c.GetFlag("V") != flags.v {
		passed = false
		errorMsg += fmt.Sprintf(" V -> exp %v, got %v ", flags.v, c.GetFlag("V"))
	}

	if passed {
		return nil
	}

	return errors.New(errorMsg)
}

func TestCPU_swabOp(t *testing.T) {
	var instruction uint16 = 000300
	tests := []struct {
		name       string
		r0Val      uint16
		swappedVal uint16
		flags      flags
		wantErr    bool
	}{
		{"swap bytes:low nonzero", 0xff00, 0x00ff, flags{false, false, false, true}, false},
		{"swap bytes:low zero", 0x0022, 0x2200, flags{false, false, true, false}, false},
	}
	for _, tt := range tests {
		c.Registers[0] = tt.r0Val
		t.Run(tt.name, func(t *testing.T) {
			swabOpcode := c.Decode(instruction)
			swabOpcode(instruction)

			// assert value:
			if c.Registers[0] != tt.swappedVal {
				t.Errorf("cpu.swapbOp r0 = %x, exp -> %x", c.Registers[0], tt.swappedVal)
			}

			// assert flags
			if err := assertFlags(tt.flags, c); err != nil {
				t.Error(err.Error())
			}
		})
	}
}
