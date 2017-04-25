package pdpcpu

import (
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

func Test_cpu_Decode(t *testing.T) {
	tests := []struct {
		name string
		c    CPU
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Decode()
		})
	}
}

func Test_cpu_Execute(t *testing.T) {
	tests := []struct {
		name string
		c    CPU
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Execute()
		})
	}
}

func Test_cpu_DumpRegisters(t *testing.T) {
	tests := []struct {
		name string
		c    CPU
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.DumpRegisters()
		})
	}
}
