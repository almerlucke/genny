package add

import "testing"

func TestAdd(t *testing.T) {
	tests := []struct{ a, x, want float64 }{{2, 3, 5}, {-1, 0.5, -0.5}, {0, 7, 7}}
	for _, tt := range tests {
		if got := New(tt.a).Shape(tt.x); got != tt.want {
			t.Errorf("Add(%v).Shape(%v) = %v want %v", tt.a, tt.x, got, tt.want)
		}
	}
}
