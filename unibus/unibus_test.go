package unibus

import (
	"testing"
)

func TestUnibus_ReadIOPage(t *testing.T) {
	type args struct {
		physicalAddress uint32
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
			got, err := u.ReadIOPage(tt.args.physicalAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unibus.ReadIOPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Unibus.ReadIOPage() = %v, want %v", got, tt.want)
			}
		})
	}
}
