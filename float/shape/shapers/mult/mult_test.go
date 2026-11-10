package mult

import "testing"

func TestMult(t *testing.T) {
	tests := []struct{ m, x, want float64 }{{2, 3, 6}, {-1, 4, -4}, {0, 9, 0}, {0.5, 3, 1.5}}
	for _, tt := range tests {
		if got := New(tt.m).Shape(tt.x); got != tt.want {
			t.Errorf("Mult(%v).Shape(%v) = %v want %v", tt.m, tt.x, got, tt.want)
		}
	}
}
