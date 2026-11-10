package mod

import "testing"

func TestMod(t *testing.T) {
	tests := []struct{ m, x, want float64 }{{3, 7, 1}, {3, 6, 0}, {3, -7, -1}, {1, 2.25, 0.25}, {0.5, 1.75, 0.25}}
	for _, tt := range tests {
		if got := New(tt.m).Shape(tt.x); got != tt.want {
			t.Errorf("Mod(%v).Shape(%v) = %v want %v", tt.m, tt.x, got, tt.want)
		}
	}
}
