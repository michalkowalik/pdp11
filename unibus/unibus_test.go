package unibus

import (
	"testing"
)

func TestUnibus_ReadIO(t *testing.T) {
	type args struct {
		physicalAddress Uint18
		byteFlag        bool
	}
	tests := []struct {
		name    string
		args    args
		want    uint16
		wantErr bool
	}{
		{"Get PSW", args{PSWAddr, false}, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u.Psw.Set(tt.want)
			got := u.ReadIO(tt.args.physicalAddress)
			if got != tt.want {
				t.Errorf("Unibus.ReadIO() = %v, want %v", got, tt.want)
			}
		})
	}
}
