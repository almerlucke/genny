package linear

import "testing"

func TestLinear(t *testing.T) {
	tests := []struct {
		name string
		l    *Linear
		x    float64
		want float64
	}{
		{"scale shift", New(2, 1), 3, 7},
		{"identity", New(1, 0), 4, 4},
		{"bipolar 0", NewBipolar(), 0, -1},
		{"bipolar .5", NewBipolar(), 0.5, 0},
		{"bipolar 1", NewBipolar(), 1, 1},
		{"unipolar -1", NewUnipolar(), -1, 0},
		{"unipolar 0", NewUnipolar(), 0, 0.5},
		{"unipolar 1", NewUnipolar(), 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.Shape(tt.x); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
