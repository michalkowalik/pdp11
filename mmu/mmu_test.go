package mmu

import "testing"

func TestMMU_ReadMemoryWord(t *testing.T) {
	type args struct {
		addr uint16
	}
	tests := []struct {
		name string
		m    *MMU
		args args
		want uint16
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.ReadMemoryWord(tt.args.addr); got != tt.want {
				t.Errorf("MMU.ReadMemoryWord() = %v, want %v", got, tt.want)
			}
		})
	}
}
