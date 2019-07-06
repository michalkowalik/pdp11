package unibus

import (
	"fmt"
	"os"
	"testing"
)

var (
	rk11 *RK11
)

func TestRK11_Attach(t *testing.T) {
	type args struct {
		drive int
		path  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"non existing file", args{0, "foo.bar.rk5"}, true},
		{"exisiting file", args{0, "../images/rk0.img"}, false},
		{"invalid drive number", args{8, "../images/rk0.img"}, true},
	}
	rk11 = NewRK(nil)
	wd, _ := os.Getwd()
	fmt.Println("Current test dir: " + wd)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := rk11.Attach(tt.args.drive, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("RK11.Attach() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
