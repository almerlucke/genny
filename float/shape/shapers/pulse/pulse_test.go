package pulse

import "testing"

func TestPulse(t *testing.T) {
	p := New(0.4)
	tests := []struct{ x, want float64 }{{0, 1}, {0.39, 1}, {0.4, 0}, {0.5, 0}, {1, 0}, {-1, 1}}
	for _, tt := range tests {
		if got := p.Shape(tt.x); got != tt.want {
			t.Errorf("Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
	p.W = 0.6
	if p.Shape(0.5) != 1 {
		t.Error("changed width")
	}
}
