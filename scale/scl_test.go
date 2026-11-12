package scale_test

import (
	"math"
	"strings"
	"testing"

	"github.com/almerlucke/genny/scale"
)

func parse(t *testing.T, text string) scale.Scl {
	t.Helper()
	s, err := scale.ParseScl(strings.NewReader(text))
	if err != nil {
		t.Fatalf("ParseScl: %v", err)
	}
	return s
}

func near(a, b, eps float64) bool { return math.Abs(a-b) <= eps }

func TestParseSclBuiltIns(t *testing.T) {
	for _, c := range []struct {
		name   string
		text   string
		notes  int
		period float64
	}{
		{"12tet", scale.SclTET12, 12, 1200},
		{"just", scale.SclJust5, 12, 1200},
		{"meantone", scale.SclMeantone, 12, 1200},
		{"pythagorean", scale.SclPythagorean, 12, 1200},
		{"bohlen-pierce", scale.SclBohlenPierce, 13, 1901.955},
	} {
		s := parse(t, c.text)
		if len(s.Cents) != c.notes || !near(s.Period(), c.period, 0.001) {
			t.Errorf("%s: %d notes, period %v", c.name, len(s.Cents), s.Period())
		}
		sc := s.Scale(60)
		if sc.Len() != c.notes || sc.Root() != 60 {
			t.Errorf("%s: scale has %d degrees", c.name, sc.Len())
		}
		if got := sc.Degree(sc.Len()); !near(got, 60+c.period/100, 1e-4) {
			t.Errorf("%s: a period up is %v", c.name, got)
		}
	}
}

func TestSclMatchesKnownTunings(t *testing.T) {
	tet := parse(t, scale.SclTET12).Scale(0)
	for i := range 12 {
		if !near(tet.Degree(i), float64(i), 1e-9) {
			t.Errorf("12-TET degree %d = %v", i, tet.Degree(i))
		}
	}
	just := parse(t, scale.SclJust5).Scale(0)
	if !near(just.Degree(7), 12*math.Log2(1.5), 1e-9) || !near(just.Degree(4), 12*math.Log2(5.0/4), 1e-9) {
		t.Errorf("just fifth %v third %v", just.Degree(7), just.Degree(4))
	}
	meantone := parse(t, scale.SclMeantone).Scale(0)
	// the meantone fifth is 696.578 cents, the third is pure
	if !near(meantone.Degree(7), 6.96578, 1e-4) || !near(meantone.Degree(4), 12*math.Log2(5.0/4), 1e-4) {
		t.Errorf("meantone fifth %v third %v", meantone.Degree(7), meantone.Degree(4))
	}
	pyth := parse(t, scale.SclPythagorean).Scale(0)
	if !near(pyth.Degree(7), 12*math.Log2(1.5), 1e-9) || !near(pyth.Degree(1), 12*math.Log2(256.0/243), 1e-9) {
		t.Errorf("pythagorean %v %v", pyth.Degree(7), pyth.Degree(1))
	}
	bp := parse(t, scale.SclBohlenPierce).Scale(0)
	if !near(bp.Degree(13), 12*math.Log2(3), 1e-9) || !near(bp.Degree(26), 24*math.Log2(3), 1e-9) {
		t.Errorf("tritave %v", bp.Degree(13))
	}
}

func TestSclScaleQuantizes(t *testing.T) {
	bp := parse(t, scale.SclBohlenPierce).Scale(48)
	for _, x := range []float64{48, 50.3, 53.7, 60, 70.1} {
		q := bp.Quantize(x)
		if bp.Quantize(q) != q {
			t.Errorf("Quantize(%v) = %v is not stable", x, q)
		}
		if math.Abs(q-x) > bp.Degree(1)-48 {
			t.Errorf("Quantize(%v) = %v is too far", x, q)
		}
	}
}

func TestParseSclFormat(t *testing.T) {
	s := parse(t, "! comment\n! another\n  A test \n! mid\n 3\n!\n 100.0 first\n 5/4  major third\r\n  2\n")
	if s.Description != "A test" {
		t.Errorf("description %q", s.Description)
	}
	if len(s.Cents) != 3 || s.Cents[0] != 100 || !near(s.Cents[1], 386.3137, 1e-3) || s.Cents[2] != 1200 {
		t.Errorf("cents %v", s.Cents)
	}
	if got := s.Steps(); len(got) != 3 || got[0] != 0 || got[1] != 1 {
		t.Errorf("steps %v", got)
	}
	// an empty description line, and extra lines after the pitches
	s = parse(t, "\n1\n2/1\nignored\n")
	if s.Description != "" || len(s.Cents) != 1 || s.Cents[0] != 1200 {
		t.Errorf("empty description: %+v", s)
	}
	// a single-note scale has just the root and the period
	if sc := s.Scale(60); sc.Len() != 1 || sc.Degree(1) != 72 {
		t.Errorf("one note scale: %d degrees, %v", sc.Len(), sc.Degree(1))
	}
}

func TestParseSclErrors(t *testing.T) {
	for name, text := range map[string]string{
		"empty":             "",
		"only comments":     "! a\n! b\n",
		"no count":          "desc\n",
		"bad count":         "desc\nx\n",
		"zero count":        "desc\n0\n",
		"negative count":    "desc\n-1\n",
		"huge count":        "desc\n999999999\n",
		"too few pitches":   "desc\n3\n100.0\n200.0\n",
		"not a number":      "desc\n1\nfoo\n",
		"bad cents":         "desc\n1\n1.2.3\n",
		"nan cents":         "desc\n1\nnan.\n",
		"zero ratio":        "desc\n1\n0/1\n",
		"zero denominator":  "desc\n1\n3/0\n",
		"negative ratio":    "desc\n1\n-3/2\n",
		"ratio of garbage":  "desc\n1\n3/x\n",
		"descending":        "desc\n2\n200.0\n100.0\n",
		"repeated":          "desc\n2\n100.0\n100.0\n",
		"negative cents":    "desc\n1\n-5.0\n",
		"unison first":      "desc\n1\n1/1\n",
		"overflowing ratio": "desc\n1\n99999999999999999999999999/1\n",
	} {
		if _, err := scale.ParseScl(strings.NewReader(text)); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func FuzzParseScl(f *testing.F) {
	for _, s := range []string{
		scale.SclTET12, scale.SclJust5, scale.SclMeantone, scale.SclPythagorean, scale.SclBohlenPierce,
		"", "!\n", "desc\n1\n2\n", "d\n2\n1.0\n2/1\n", "d\n1\n-1.0\n", "d\n1\n1e400.\n", "d\n99\n", "\n\n\n",
		"d\n1\n0/0\n", "d\n2\n5/4 x\n 3/2 y\n", "\xff\xfe\n1\n1.5\n",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, text string) {
		s, err := scale.ParseScl(strings.NewReader(text))
		if err != nil {
			return
		}
		if len(s.Cents) == 0 {
			t.Fatal("no error but no pitches")
		}
		prev := 0.0
		for _, c := range s.Cents {
			if !(c > prev) || math.IsInf(c, 0) {
				t.Fatalf("pitches do not ascend: %v", s.Cents)
			}
			prev = c
		}
		sc := s.Scale(60) // must not panic
		_ = sc.Quantize(61.5)
	})
}
