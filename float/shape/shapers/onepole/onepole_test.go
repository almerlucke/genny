package onepole

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/shape"
)

var _ shape.Shaper = (*Filter)(nil)

func TestLowpassStep(t *testing.T) {
	f := New(0.5)
	want := []float64{0.5, 0.75, 0.875, 0.9375}
	for i, w := range want {
		if got := f.Shape(1); math.Abs(got-w) > 1e-12 {
			t.Errorf("step %d: got %v want %v", i, got, w)
		}
	}
	f.Reset()
	if got := f.Shape(1); got != 0.5 {
		t.Errorf("after Reset got %v", got)
	}
}

func TestHighpassStep(t *testing.T) {
	f := NewHighpass(0.5)
	want := []float64{0.5, 0.25, 0.125}
	for i, w := range want {
		if got := f.Shape(1); math.Abs(got-w) > 1e-12 {
			t.Errorf("step %d: got %v want %v", i, got, w)
		}
	}
}

func TestLowPlusHighIsInput(t *testing.T) {
	lo, hi := New(0.2), NewHighpass(0.2)
	for i := 0; i < 200; i++ {
		x := math.Sin(float64(i) / 5)
		if s := lo.Shape(x) + hi.Shape(x); math.Abs(s-x) > 1e-12 {
			t.Fatalf("lowpass + highpass = %v want %v", s, x)
		}
	}
}

func TestTimeConstant(t *testing.T) {
	// after tau samples the step response is 1-1/e
	tau := 100.0
	f := NewSamples(tau)
	var y float64
	for i := 0; i < 100; i++ {
		y = f.Shape(1)
	}
	if want := 1 - math.Exp(-1); math.Abs(y-want) > 0.01 {
		t.Errorf("step after tau = %v want %v", y, want)
	}
	g := NewTime(0.01, 10000) // 100 samples
	if math.Abs(g.a-f.a) > 1e-15 {
		t.Errorf("NewTime coefficient %v want %v", g.a, f.a)
	}
	if NewSamples(0).a != 1 || NewTime(0, 44100).a != 1 {
		t.Error("zero time constant passes the input")
	}
	if a := Coefficient(1e300); !(a > 0 && a <= 1) {
		t.Errorf("huge tau coefficient %v", a)
	}
	if hp := NewHighpassTime(0.01, 10000); !hp.highpass || hp.a != f.a {
		t.Error("NewHighpassTime")
	}
}

func TestStableAndBounded(t *testing.T) {
	for _, a := range []float64{1e-9, 0.01, 0.5, 1} {
		lo, hi := New(a), NewHighpass(a)
		for i := 0; i < 5000; i++ {
			x := 1.0
			if i%7 < 3 {
				x = -1
			}
			l, h := lo.Shape(x), hi.Shape(x)
			if math.Abs(l) > 1+1e-12 || math.Abs(h) > 2 || math.IsNaN(l+h) {
				t.Fatalf("a=%v unstable: %v %v", a, l, h)
			}
		}
	}
	f := New(0.3)
	f.Shape(0.5)
	prev := f.Shape(0.5)
	for _, x := range []float64{math.NaN(), math.Inf(1)} {
		if f.Shape(x) != prev {
			t.Errorf("non finite %v must be ignored", x)
		}
	}
}

func TestPanics(t *testing.T) {
	tests := map[string]func(){
		"a 0":          func() { New(0) },
		"a >1":         func() { New(1.5) },
		"a NaN":        func() { New(math.NaN()) },
		"hp a 0":       func() { NewHighpass(0) },
		"tau <0":       func() { NewSamples(-1) },
		"tau NaN":      func() { Coefficient(math.NaN()) },
		"seconds < 0":  func() { NewTime(-1, 44100) },
		"rate 0":       func() { NewTime(1, 0) },
		"rate NaN":     func() { NewTime(1, math.NaN()) },
		"hp rate":      func() { NewHighpassTime(1, -1) },
		"seconds +Inf": func() { NewTime(math.Inf(1), 1) },
	}
	for name, f := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: expected panic", name)
				}
			}()
			f()
		}()
	}
}
