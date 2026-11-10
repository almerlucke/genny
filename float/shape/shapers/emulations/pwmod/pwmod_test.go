package pwmod

import "testing"

func TestPWMod(t *testing.T) {
	p := New()
	tests := []struct{ x, want float64 }{
		{0.2, 1},  // .25 < .4
		{0.4, -1}, // .5
		{0.6, -1}, // .75
		{0.8, 1},  // 1.0 wraps to 0
		{0.9, 1},  // 1.125 wraps to .125
	}
	for _, tt := range tests {
		if got := p.Shape(tt.x); got != tt.want {
			t.Errorf("Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
	p.SetA1(1)
	if p.Shape(0.5) != -1 {
		t.Error("a1=1 at .5")
	}
	p.SetWidth(0.6)
	if p.Shape(0.5) != 1 {
		t.Error("wider pulse")
	}
}
