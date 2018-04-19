package psw

import (
	"testing"
)

func TestGet(t *testing.T) {
	var p PSW
	p = 1

	if p.Get() != 1 {
		t.Errorf("Expected PSW value of 1, got %v", p.Get())
	}
}

func TestPSW_C(t *testing.T) {
	tests := []struct {
		name string
		p    PSW
		want bool
	}{
		{"C set, all 0", 1, true},
		{"C set, other flags too", 3, true},
		{"C clear, all 0", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.p
			if p.C() != tt.want {
				t.Errorf("pws.C() (%v) failed. P: %v, wanted %v, got %v",
					tt.name, p, tt.want, p.C())
			}
		})
	}
}

func TestPSW_SetC(t *testing.T) {
	type args struct {
		status bool
	}
	tests := []struct {
		name string
		psw  PSW
		args bool
	}{
		{"set C P=0", 0, true},
		{"set C P=1", 1, true},
		{"clear C P=0", 0, false},
		{"clear C P=1", 1, false},
		{"clear C P=3", 3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &(tt.psw)
			p.SetC(tt.args)
			if p.C() != tt.args {
				t.Errorf("psw.SetC() (%s) failed. P: %v, expected: %v, got %v",
					tt.name, tt.psw, tt.args, p.C())
			}
		})
	}
}

func TestPSW_SetN(t *testing.T) {
	type args struct {
		status bool
	}
	tests := []struct {
		name        string
		psw         PSW
		args        bool
		modifiedPsw PSW
	}{
		{"set N P=0", 0, true, 8},
		{"set N P=8", 8, true, 8},
		{"clear N P=0", 0, false, 0},
		{"clear N P=8", 8, false, 0},
		{"clear N P=9", 9, false, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.psw.SetN(tt.args)
			if tt.psw.N() != tt.args {
				t.Errorf("psw.SetN() (%s) failed. P: %v, expected: %v, got %v \n",
					tt.name, tt.psw, tt.args, tt.psw.N())
			}

			if tt.psw != tt.modifiedPsw {
				t.Errorf("psw.SetN (%s) failed. P = %v, expected P = %v\n",
					tt.name, tt.psw, tt.modifiedPsw)
			}
		})
	}
}
