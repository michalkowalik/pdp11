package unibus

import (
	"pdp/console"
	"pdp/psw"
	"testing"
)

func Test_cpu_Fetch(t *testing.T) {
	tests := []struct {
		name string
		c    CPU
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Fetch()
		})
	}
}

func TestCPU_GetFlag(t *testing.T) {
	p := psw.PSW(0)
	var cons console.Console = console.NewSimple()
	u := New(&p, nil, &cons, false)

	var c = NewCPU(u.Mmu, u, false)

	tests := []struct {
		name       string
		statusWord uint16
		args       string
		want       bool
	}{
		{"C unset", 0, "C", false},
		{"C set", 1, "C", true},
		{"C and V set", 3, "C", true},
		{"V unset", 1, "V", false},
		{"V set", 3, "V", true},
		{"Z set", 4, "Z", true},
		{"Z set, C unset", 4, "C", false},
		{"Z set, N unset", 4, "N", false},
		{"N Set", 8, "N", true},
		{"T unset", 0xf, "T", false},
		{"T set", 0x1f, "T", true},
	}
	for _, tt := range tests {
		tempPsw := psw.PSW(tt.statusWord)
		c.unibus.Psw = &tempPsw
		t.Run(tt.name, func(t *testing.T) {
			if got := c.GetFlag(tt.args); got != tt.want {
				t.Errorf("CPU.GetFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}
