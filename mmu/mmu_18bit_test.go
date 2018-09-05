package mmu

import (
	"testing"
)

func TestMMU18Bit_mapVirtualToPhysical(t *testing.T) {
	type args struct {
		virtualAddress uint16
		psw            uint16
	}
	tests := []struct {
		name    string
		m       *MMU18Bit
		args    args
		want    uint32
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.mapVirtualToPhysical(tt.args.virtualAddress)
			if got != tt.want {
				t.Errorf("MMU18Bit.mapVirtualToPhysical() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMMU18Bit_ReadMemoryWord(t *testing.T) {
	type args struct {
		addr uint16
	}
	tests := []struct {
		name string
		m    *MMU18Bit
		args args
		want uint16
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.m.ReadMemoryWord(tt.args.addr); got != tt.want {
				t.Errorf("MMU18Bit.ReadMemoryWord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMMU18Bit_WriteMemoryByte(t *testing.T) {
	type args struct {
		addr uint16
		data byte
	}
	tests := []struct {
		name    string
		m       *MMU18Bit
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.WriteMemoryByte(tt.args.addr, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("MMU18Bit.WriteMemoryByte() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
