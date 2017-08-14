package pdpcpu

import (
	"os"
	"pdp/mmu"
	"testing"
)

// global shared resources: CPU, memory etc.
var c *CPU
var memory [0x400000]byte // 64KB of memory is all everyone needs

// TestMain to resucure -> initialize memory and CPU
func TestMain(m *testing.M) {
	mmu := &mmu.MMU{}
	mmu.Memory = &memory
	c = New(mmu)

	os.Exit(m.Run())
}

func TestCPU_clrOp(t *testing.T) {
	type args struct {
		instruction int16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"value in register", args{05000}, false},
		{"address in register", args{05011}, false},
	}
	c.Registers[0] = 0xff
	c.Registers[1] = 0xff
	c.mmunit.Memory[0xff] = 2

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.clrOp(tt.args.instruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("CPU.clrOp() error = %v, wantErr %v", err, tt.wantErr)
			}
			// also: check if value is really 0:
			op := uint16(tt.args.instruction) & 077
			t.Logf("instruction: %x, op: %x\n", tt.args.instruction, op)
			w := c.readWord(op)
			if w != 0 {
				t.Errorf("CPU.clrOp() -> destination for %v is set to %x", op, w)
			}
		})
	}
}

func TestCPU_addOp(t *testing.T) {
	type args struct {
		instruction int16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantRes int16
	}{
		{"dst in register, src memory value", args{061100}, false, 0x1fe},
	}

	c.Registers[0] = 0xff
	c.Registers[1] = 0xff
	c.Registers[2] = 0
	c.Registers[3] = 2
	c.mmunit.Memory[0xff] = 0xff
	c.mmunit.Memory[0] = 2
	c.mmunit.Memory[3] = 3

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.addOp(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("CPU.addOp() error = %v, wantErr %v", err, tt.wantErr)
			}
			// also -> check value
			w := c.readWord(uint16(tt.args.instruction & 077))
			t.Logf("Value at dst: %x\n", w)
			if int16(w) != tt.wantRes {
				t.Errorf("expected %x, got %x", tt.wantRes, w)
			}
		})
	}
}

func TestCPU_movOp(t *testing.T) {
	type args struct {
		instruction int16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		dst     int16
	}{
		{"move from memory to register", args{011001}, false, 4},
	}

	c.Registers[0] = 0xff
	c.Registers[1] = 0
	c.mmunit.Memory[0xff] = uint8(4)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.movOp(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("CPU.movOp() error = %v, wantErr %v", err, tt.wantErr)
			}
			d := c.readWord(uint16(tt.args.instruction & 077))

			if int16(d) != tt.dst {
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
		instruction int16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		dst     uint16
	}{
		{"complement dst on value in register", args{005100}, false, 0xff0f},
	}

	c.Registers[0] = 0xf0
	c.mmunit.Memory[0xff] = uint8(4)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opcode := c.Decode(uint16(tt.args.instruction))
			if err := opcode(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("CPU.comOp() error = %v, wantErr %v", err, tt.wantErr)
			}
			d := c.readWord(uint16(tt.args.instruction & 077))

			if d != tt.dst {
				t.Logf("destination addr: %x\n", tt.args.instruction&077)
				t.Errorf("Expected destination: %x, got %x\n", tt.dst, d)
			}
		})
	}
}

func TestCPU_incOp(t *testing.T) {
	type args struct {
		instruction int16
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
			args{05200}, 0x7FFF, 0x8000, false, true, false, true},
		{"INC on 0x0000 should set no flags",
			args{05200}, 0, 1, false, false, false, false},
		{"INC on 0xffff should set Z flag",
			args{05200}, 0xffff, 0, false, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Registers[0] = tt.regVal
			instruction := c.Decode(uint16(tt.args.instruction))
			if err := instruction(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("CPU.incOp() error = %v, wantErr %v", err, tt.wantErr)
			}

			d := c.readWord(uint16(tt.args.instruction & 077))
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
		instruction int16
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
			if err := instruction(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("CPU.negOp() error = %v, wantErr %v", err, tt.wantErr)
			}
			if c.Registers[0] != tt.dst {
				t.Errorf("NEG returned unexpected result. expected %v, got %v\n",
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

func TestCPU_adcOp(t *testing.T) {
	type args struct {
		instruction int16
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
			if err := instruction(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("CPU.adcOp() error = %v, wantErr %v", err, tt.wantErr)
			}
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
		instruction int16
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantRes uint16
	}{
		{"dst value in REG", args{074002}, false, 000325},
	}

	c.Registers[0] = 001234
	c.Registers[2] = 001111

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.xorOp(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("CPU.xorOp() error = %v, wantErr %v", err, tt.wantErr)
			}
			w := c.readWord(uint16(tt.args.instruction & 077))
			t.Logf("Value at dst: %x \n", w)
			if w != tt.wantRes {
				t.Errorf("expected %x, got %x\n", tt.wantRes, w)
			}
		})
	}
}

func TestCPU_ashcOp(t *testing.T) {
	type args struct {
		instruction int16
	}
	tests := []struct {
		name    string
		c       *CPU
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.c.ashcOp(tt.args.instruction); (err != nil) != tt.wantErr {
				t.Errorf("CPU.ashcOp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
