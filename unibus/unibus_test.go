package unibus

import (
	"pdp/psw"
	"testing"
)

func TestUnibus_ReadIO(t *testing.T) {
	type args struct {
		physicalAddress Uint18
		byteFlag        bool
	}
	tests := []struct {
		name string
		args args
		want uint16
	}{
		{"Get PSW", args{PSWAddr, false}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*u.Psw = psw.PSW(tt.want)
			got := u.ReadIO(tt.args.physicalAddress)
			if got != tt.want {
				t.Errorf("Unibus.ReadIO() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnibus_PSWStatus(t *testing.T) {

	tests := []struct {
		name            string
		initialPSWValue uint16
		currentMode     uint16
		newMode         uint16
	}{
		{"kernel=>kernel", 0, KernelMode, KernelMode},
		{"user=>user", 0140000, UserMode, UserMode},
		{"kernel=>user", 0, KernelMode, UserMode},
		{"user=>kernel", 0140000, UserMode, KernelMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u.Psw.Set(tt.initialPSWValue)
			if tt.currentMode != u.Psw.GetMode() {
				t.Errorf("current mode doesn't match. expected %d, got %d",
					tt.currentMode, u.Psw.GetMode())
			}
			u.PdpCPU.SwitchMode(tt.newMode)

			if tt.newMode != u.Psw.GetMode() {
				t.Errorf("Expected PSW mode doesn't match. expected %d, got %d",
					tt.newMode, u.Psw.GetMode())
			}
			if tt.currentMode != u.Psw.GetPreviousMode() {
				t.Errorf("Expected psw previousMode doesn't macht. expected %d, got %d",
					tt.currentMode, u.Psw.GetPreviousMode())
			}
		})
	}
}
