package clip

import (
	"math"
	"testing"
)

func TestHardClip(t *testing.T) {
	c := New(Hard, -0.5, 0.75)
	tests := []struct{ x, want float64 }{{0.2, 0.2}, {-0.5, -0.5}, {0.75, 0.75}, {1, 0.75}, {-3, -0.5}, {0, 0}}
	for _, tt := range tests {
		if got := c.Shape(tt.x); got != tt.want {
			t.Errorf("Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
}

func TestSoftClip(t *testing.T) {
	for _, a := range []float64{1, 2, 10} {
		c := New(Soft, a)
		// normalised: +-1 map to +-1, 0 to 0, odd, monotone
		if math.Abs(c.Shape(1)-1) > 1e-12 || math.Abs(c.Shape(-1)+1) > 1e-12 || c.Shape(0) != 0 {
			t.Fatalf("a=%v: 1 -> %v, -1 -> %v, 0 -> %v", a, c.Shape(1), c.Shape(-1), c.Shape(0))
		}
		prev := math.Inf(-1)
		for x := -2.0; x <= 2; x += 0.1 {
			v := c.Shape(x)
			if v <= prev {
				t.Fatalf("a=%v not increasing at %v", a, x)
			}
			if math.Abs(c.Shape(-x)+v) > 1e-12 {
				t.Fatalf("a=%v not odd at %v", a, x)
			}
			prev = v
		}
	}
	// a=1: atan(0.5) / atan(1)
	if got, want := New(Soft, 1).Shape(0.5), math.Atan(0.5)/(math.Pi/4); math.Abs(got-want) > 1e-12 {
		t.Fatalf("got %v want %v", got, want)
	}
	// a large input is squeezed below atan(inf)/atan(a)
	if got, limit := New(Soft, 2).Shape(1e9), (math.Pi/2)/math.Atan(2); got >= limit || got < limit-1e-6 {
		t.Fatalf("got %v, limit %v", got, limit)
	}
}
