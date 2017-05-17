package pdpcpu

import (
	"pdp/mmu"
	"testing"
)

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

	var c = &CPU{}
	var memory [0x400000]byte // 64KB of memory is all everyone needs
	c.Registers[0] = 0xff
	c.Registers[1] = 0xff
	c.mmunit = &mmu.MMU{}
	c.mmunit.Memory = &memory
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

	var c = &CPU{}
	var memory [0x400000]byte // 64KB of memory is all everyone needs
	c.Registers[0] = 0xff
	c.Registers[1] = 0xff
	c.Registers[2] = 0
	c.Registers[3] = 2
	c.mmunit = &mmu.MMU{}
	c.mmunit.Memory = &memory
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
