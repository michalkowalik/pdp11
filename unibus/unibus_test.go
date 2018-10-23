package unibus

import "testing"

func TestUnibus_ReadIOPage(t *testing.T) {
	type args struct {
		physicalAddress uint32
		byteFlag        bool
	}
	tests := []struct {
		name    string
		u       *Unibus
		args    args
		want    uint16
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.u.ReadIOPage(tt.args.physicalAddress, tt.args.byteFlag)
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
