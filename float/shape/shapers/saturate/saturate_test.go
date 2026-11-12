package saturate

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var (
	_ shape.Shaper = (*Saturator)(nil)
	_ shape.Shaper = (*Tube)(nil)
)

func TestKnownValues(t *testing.T) {
	tests := []struct {
		name string
		s    shape.Shaper
		x    float64
		want float64
	}{
		{"tanh at 1", New(Tanh, 3), 1, 1},
		{"tanh at 0.5", New(Tanh, 2), 0.5, math.Tanh(1) / math.Tanh(2)},
		{"atan at 0.5", New(Atan, 1), 0.5, math.Atan(0.5) / (math.Pi / 4)},
		{"logistic at 1", New(Logistic, 5), 1, 1},
		{"logistic at 0.5", New(Logistic, 2), 0.5, (2/(1+math.Exp(-1)) - 1) / (2/(1+math.Exp(-2)) - 1)},
		{"tube bias 0 is tanh", NewTube(2, 0), 0.3, New(Tanh, 2).Shape(0.3)},
		{"tube at 0", NewTube(2, 0.4), 0, 0},
	}
	for _, tt := range tests {
		if got := tt.s.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("%s: got %v want %v", tt.name, got, tt.want)
		}
	}
}

func TestSymmetricProperties(t *testing.T) {
	for _, mode := range []Mode{Tanh, Atan, Logistic} {
		for _, drive := range []float64{0.01, 0.5, 1, 4, 50} {
			s := New(mode, drive)
			if s.Shape(0) != 0 || math.Abs(s.Shape(1)-1) > 1e-12 || math.Abs(s.Shape(-1)+1) > 1e-12 {
				t.Fatalf("mode %v drive %v: bad anchors", mode, drive)
			}
			prev := math.Inf(-1)
			for x := -2.0; x <= 2; x += 0.01 {
				y := s.Shape(x)
				if y < prev || math.Abs(s.Shape(-x)+y) > 1e-12 {
					t.Fatalf("mode %v drive %v: not monotone/odd at %v", mode, drive, x)
				}
				if x >= -1 && x <= 1 && math.Abs(y) > 1+1e-12 {
					t.Fatalf("mode %v drive %v: out of range at %v: %v", mode, drive, x, y)
				}
				prev = y
			}
			for _, x := range []float64{1e300, -1e300, math.MaxFloat64, math.Inf(-1)} {
				if y := s.Shape(x); math.IsNaN(y) || math.IsInf(y, 0) {
					t.Fatalf("mode %v drive %v: Shape(%v) = %v", mode, drive, x, y)
				}
			}
		}
	}
}

func TestTube(t *testing.T) {
	for _, bias := range []float64{-0.8, -0.2, 0, 0.3, 0.9, 0.99} {
		for _, drive := range []float64{0.1, 1, 5, 30} {
			s := NewTube(drive, bias)
			if s.Shape(0) != 0 {
				t.Fatalf("drive %v bias %v: Shape(0) = %v", drive, bias, s.Shape(0))
			}
			prev := math.Inf(-1)
			maxAbs := 0.0
			for i := 0; i <= 400; i++ {
				x := -1 + float64(i)/200
				y := s.Shape(x)
				if y < prev || math.IsNaN(y) || math.Abs(y) > 1+1e-12 {
					t.Fatalf("drive %v bias %v: bad value %v at %v", drive, bias, y, x)
				}
				maxAbs = math.Max(maxAbs, math.Abs(y))
				prev = y
			}
			if math.Abs(maxAbs-1) > 1e-9 {
				t.Fatalf("drive %v bias %v: peak %v, want 1", drive, bias, maxAbs)
			}
			if y := s.Shape(1e300); math.IsNaN(y) || math.IsInf(y, 0) {
				t.Fatalf("huge input gave %v", y)
			}
		}
	}
	// even harmonics: the curve is not odd when biased
	if s := NewTube(3, 0.5); math.Abs(s.Shape(0.5)+s.Shape(-0.5)) < 0.01 {
		t.Fatal("biased tube is odd")
	}
}

func TestPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"zero drive": func() { New(Tanh, 0) },
		"neg drive":  func() { New(Atan, -1) },
		"nan drive":  func() { New(Logistic, math.NaN()) },
		"inf drive":  func() { New(Tanh, math.Inf(1)) },
		"bad mode":   func() { New(Mode(9), 1) },
		"tube drive": func() { NewTube(0, 0) },
		"tube bias":  func() { NewTube(1, math.Inf(1)) },
		"tube big":   func() { NewTube(1, 1) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			f()
		}()
	}
}
