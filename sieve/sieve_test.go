package sieve

import (
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/sequence"
)

func TestNew(t *testing.T) {
	for _, c := range []struct {
		mod, start, lo, hi int
		want               []int
	}{
		{3, 0, 0, 10, []int{0, 3, 6, 9}},
		{3, 1, 0, 10, []int{1, 4, 7, 10}},
		{3, 4, 0, 10, []int{1, 4, 7, 10}}, // start beyond lo
		{3, -2, 0, 10, []int{1, 4, 7, 10}},
		{5, 2, 3, 20, []int{7, 12, 17}},
		{1, 0, 2, 5, []int{2, 3, 4, 5}},
		{4, 0, -8, 0, []int{-8, -4, 0}},
		{3, 0, 5, 5, nil},
		{3, 2, 5, 5, []int{5}},
		{3, 0, 10, 0, nil}, // hi < lo
	} {
		if got := New(c.mod, c.start, c.lo, c.hi).Values(); !slices.Equal(got, c.want) {
			t.Errorf("New(%d,%d,%d,%d) = %v, want %v", c.mod, c.start, c.lo, c.hi, got, c.want)
		}
	}
}

func TestNewPanics(t *testing.T) {
	for _, mod := range []int{0, -3} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("mod %d did not panic", mod)
				}
			}()
			New(mod, 0, 0, 10)
		}()
	}
}

func TestOperations(t *testing.T) {
	a := New(3, 0, 0, 12)
	b := New(2, 0, 0, 12)
	for _, c := range []struct {
		name string
		got  Sieve
		want []int
	}{
		{"union", a.Union(b), []int{0, 2, 3, 4, 6, 8, 9, 10, 12}},
		{"intersection", a.Intersection(b), []int{0, 6, 12}},
		{"difference", a.Difference(b), []int{3, 9}},
		{"complement", a.Complement(), []int{1, 2, 4, 5, 7, 8, 10, 11}},
		{"double complement", a.Complement().Complement(), []int{0, 3, 6, 9, 12}},
		{"empty union", (Sieve{}).Union(a), []int{0, 3, 6, 9, 12}},
		{"empty intersection", (Sieve{}).Intersection(a), nil},
		{"of", Of(0, 5, 4, 1, 4, 9, -1), []int{1, 4}},
	} {
		if !slices.Equal(c.got.Values(), c.want) {
			t.Errorf("%s = %v, want %v", c.name, c.got.Values(), c.want)
		}
	}
	if lo, hi := a.Union(New(1, 0, -2, 20)).Range(); lo != -2 || hi != 20 {
		t.Errorf("union range = %d..%d", lo, hi)
	}
	if lo, hi := (Sieve{}).Union(a).Range(); lo != 0 || hi != 12 {
		t.Errorf("union with empty range = %d..%d", lo, hi)
	}
}

func TestComplementOfEmptyRange(t *testing.T) {
	if n := New(2, 0, 5, 1).Complement().Len(); n != 0 {
		t.Errorf("complement of an empty range has %d members", n)
	}
	if n := (Sieve{}).Complement().Len(); n != 0 {
		t.Errorf("complement of the zero sieve has %d members", n)
	}
}

func TestContains(t *testing.T) {
	s := New(4, 1, 0, 20)
	for x, want := range map[int]bool{1: true, 5: true, 21: false, 2: false, -3: false, 17: true} {
		if s.Contains(x) != want {
			t.Errorf("Contains(%d) = %v", x, !want)
		}
	}
}

func TestImmutable(t *testing.T) {
	s := New(2, 0, 0, 6)
	v := s.Values()
	v[0] = 99
	if s.Values()[0] != 0 {
		t.Error("Values exposes the sieve")
	}
}

func TestScale(t *testing.T) {
	got := Scale(New(5, 0, 0, 12).Union(New(7, 0, 0, 12)), 60)
	want := []float64{60, 65, 67, 70}
	if !slices.Equal(got, want) {
		t.Errorf("Scale = %v, want %v", got, want)
	}
	if len(Scale(Sieve{}, 60)) != 0 {
		t.Error("empty sieve has pitches")
	}
}

func TestGen(t *testing.T) {
	g := NewGen(New(3, 0, 0, 9), false)
	var got []int
	for range 6 {
		got = append(got, g.Generate())
	}
	if !slices.Equal(got, []int{0, 3, 6, 9, 9, 9}) || !g.Done() || g.Continuous() {
		t.Errorf("once = %v done %v", got, g.Done())
	}
	g.Reset()
	if g.Done() || g.Generate() != 0 {
		t.Error("reset")
	}

	l := NewGen(New(3, 0, 0, 6), true)
	got = nil
	for range 7 {
		got = append(got, l.Generate())
	}
	if !slices.Equal(got, []int{0, 3, 6, 0, 3, 6, 0}) || l.Done() || !l.Continuous() {
		t.Errorf("loop = %v", got)
	}
}

func TestGenEmpty(t *testing.T) {
	g := NewGen(Sieve{}, true)
	if !g.Done() || g.Generate() != 0 {
		t.Error("empty gen should be done and generate 0")
	}
	g.Reset()
	if !g.Done() {
		t.Error("empty gen done after reset")
	}
}

func TestGate(t *testing.T) {
	g := NewGate(New(3, 0, 0, 100), 8)
	var got []bool
	for range 16 {
		got = append(got, g.Generate())
	}
	want := []bool{true, false, false, true, false, false, true, false}
	if !slices.Equal(got[:8], want) || !slices.Equal(got[8:], want) {
		t.Errorf("gate = %v", got)
	}
	g.Reset()
	if !g.Generate() {
		t.Error("reset")
	}
	// a short range is false beyond it
	g = NewGate(New(1, 0, 0, 1), 4)
	got = nil
	for range 4 {
		got = append(got, g.Generate())
	}
	if !slices.Equal(got, []bool{true, true, false, false}) {
		t.Errorf("short range gate = %v", got)
	}
	if !g.Continuous() || g.Done() {
		t.Error("gate loops forever")
	}
}

func TestGatePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("length 0 did not panic")
		}
	}()
	NewGate(Sieve{}, 0)
}

func TestFilterAndSelect(t *testing.T) {
	f := NewFilter[int](sequence.NewLoop(1, 2, 3, 4, 5, 6), New(2, 1, 0, 5), 6)
	var got []int
	for range 6 {
		o := f.Generate()
		if o.Played() {
			got = append(got, o.Value)
		} else {
			got = append(got, 0)
		}
	}
	if !slices.Equal(got, []int{0, 2, 0, 4, 0, 6}) {
		t.Errorf("filter = %v", got)
	}
	if got := Select([]string{"a", "b", "c", "d"}, New(3, 0, 0, 20)); !slices.Equal(got, []string{"a", "d"}) {
		t.Errorf("select = %v", got)
	}
}

func TestFunnel(t *testing.T) {
	genny.Seed(1)
	allowed := []float64{0, 4, 7, 12}
	for _, c := range []struct{ in, want float64 }{
		{0, 0}, {1, 0}, {3, 4}, {5, 4}, {6, 7}, {100, 12}, {-5, 0}, {10, 12},
	} {
		if got := Funnel([]float64{c.in}, allowed)[0]; got != c.want {
			t.Errorf("Funnel(%v) = %v, want %v", c.in, got, c.want)
		}
	}
	// ties pick either side, in both ways
	seen := map[float64]bool{}
	for range 50 {
		seen[Funnel([]float64{2}, allowed)[0]] = true
		seen[Funnel([]float64{2}, []float64{4, 0, 4})[0]] = true
	}
	if len(seen) != 2 || !seen[0] || !seen[4] {
		t.Errorf("tie results = %v", seen)
	}
	// unsorted and duplicate allowed values, no allowed values
	if got := Funnel([]float64{6.9}, []float64{12, 7, 0, 7}); got[0] != 7 {
		t.Errorf("unsorted = %v", got)
	}
	if got := Funnel([]float64{1.5, 2}, nil); !slices.Equal(got, []float64{1.5, 2}) {
		t.Errorf("no allowed = %v", got)
	}
	if len(Funnel(nil, allowed)) != 0 {
		t.Error("no values")
	}
}

func TestInterpret(t *testing.T) {
	p := []float64{1, 1, 1, 1, 1, 1}
	s := Of(0, 5, 0, 3)
	for _, c := range []struct {
		kind int
		want []float64
	}{
		{1, []float64{1, -1, -1, 1, -1, -1}},
		{2, []float64{3, 3}},
		{3, []float64{-1, 1, 1, -1, 1, 1}},
		{4, []float64{-1, 1, 2, 1, 1}}, // leading rest; 3 is absorbed by 2
	} {
		got := Interpret(p, s, c.kind)
		if !slices.Equal(got, c.want) {
			t.Errorf("kind %d = %v, want %v", c.kind, got, c.want)
		}
	}
}

func TestInterpretKeepsTotal(t *testing.T) {
	p := []float64{1, 0.5, -0.5, 2, 1, 1, -1}
	total := 0.0
	for _, v := range p {
		total += abs(v)
	}
	for kind := 1; kind <= 4; kind++ {
		for _, s := range []Sieve{New(2, 0, 0, 6), New(3, 1, 0, 6), Sieve{}, New(1, 0, 0, 6)} {
			sum := 0.0
			for _, v := range Interpret(p, s, kind) {
				sum += abs(v)
			}
			if sum != total {
				t.Errorf("kind %d: total %v, want %v", kind, sum, total)
			}
		}
	}
}

func TestInterpretRests(t *testing.T) {
	// a member that is a rest stays a rest when it absorbs
	got := Interpret([]float64{-1, 1, 1}, Of(0, 2, 0), 2)
	if !slices.Equal(got, []float64{-3}) {
		t.Errorf("= %v", got)
	}
	// no members: kind 2 gives one rest
	got = Interpret([]float64{1, 1}, Sieve{}, 2)
	if !slices.Equal(got, []float64{-2}) {
		t.Errorf("no members = %v", got)
	}
	if got := Interpret(nil, Sieve{}, 1); len(got) != 0 {
		t.Errorf("empty pattern = %v", got)
	}
}

func TestInterpretPanics(t *testing.T) {
	for _, kind := range []int{0, 5, -1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("kind %d did not panic", kind)
				}
			}()
			Interpret([]float64{1}, Sieve{}, kind)
		}()
	}
}
