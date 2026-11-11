package constrain_test

import (
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constrain"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/sequence"
)

func take[T any](g genny.Generator[T], n int) (out []T) {
	for range n {
		out = append(out, g.Generate())
	}
	return
}

// random pitches 50..79
func pitches() genny.Generator[int] {
	return function.New(func() int { return 50 + genny.Intn(30) })
}

func absInt(a int) int { return int(math.Abs(float64(a))) }

func TestAllowAndBlockInterval(t *testing.T) {
	genny.Seed(1)
	allow := take(constrain.AllowInterval(pitches(), []int{1, 2, 5}), 300)
	for i := 1; i < len(allow); i++ {
		if d := absInt(allow[i] - allow[i-1]); d != 1 && d != 2 && d != 5 {
			t.Fatalf("step %d: interval %d", i, d)
		}
	}
	block := take(constrain.BlockInterval(pitches(), []int{0, 1, 11}), 300)
	for i := 1; i < len(block); i++ {
		if d := absInt(block[i] - block[i-1]); d == 0 || d == 1 || d == 11 {
			t.Fatalf("step %d: interval %d", i, d)
		}
	}
}

func TestAllowIntervalPitchClass(t *testing.T) {
	// 60, then 61 (1), 62 (2) and 74 (14 = 2 mod 12). Allowed is the interval 2
	run := func(opts ...constrain.Option) (out []int) {
		g := constrain.AllowInterval(sequence.New(60, 61, 62, 74), []int{2}, opts...)
		for !g.Done() {
			out = append(out, g.Generate())
		}
		return
	}
	// plain: 60, 62 (61 skipped), then 74 is 12 away and is the last value, so it is used
	if got := run(); !slices.Equal(got, []int{60, 62, 74}) {
		t.Errorf("plain: %v", got)
	}
	// with a limit of one draw nothing is skipped
	if got := run(constrain.Stop(1)); !slices.Equal(got, []int{60, 61, 62, 74}) {
		t.Errorf("stop 1: %v", got)
	}
	// the candidates after 60 are 74 and then 62: 74 is a 2 mod 12 away, so with PitchClass it is taken
	seq := func() genny.Generator[int] { return sequence.NewLoop(60, 74, 62) }
	if got := take(constrain.AllowInterval(seq(), []int{2}, constrain.PitchClass()), 2); got[1] != 74 {
		t.Errorf("pitch class: %v", got)
	}
	if got := take(constrain.AllowInterval(seq(), []int{2}), 2); got[1] != 62 {
		t.Errorf("plain: %v", got)
	}
}

func TestIntervalFloats(t *testing.T) {
	g := constrain.AllowInterval(sequence.New(60.0, 60.4, 62.0), []int{2}, constrain.Stop(5))
	got := []float64{g.Generate(), g.Generate(), g.Generate()}
	if got[0] != 60 || got[1] != 62 {
		// 60.4 is 0 away (rounded), not allowed; 62 is
		t.Errorf("%v", got)
	}
}

func TestInfeasibleHitsRetryLimit(t *testing.T) {
	calls := 0
	src := function.New(func() int { calls++; return 7 })
	g := constrain.RemoveRepetitions(src, constrain.Stop(5))
	g.Generate() // the first is always allowed
	if calls != 1 {
		t.Fatalf("first value drew %d times", calls)
	}
	if v := g.Generate(); v != 7 { // gives up and returns the candidate
		t.Errorf("got %d", v)
	}
	if calls != 6 {
		t.Errorf("%d draws, want 1+5", calls)
	}

	calls = 0
	none := constrain.AllowInterval(function.New(func() int { calls++; return calls }), nil)
	take(none, 3) // first, then two that can never be allowed
	if calls != 1+2*constrain.DefaultStop {
		t.Errorf("default stop: %d draws", calls)
	}
}

func TestEmptyInput(t *testing.T) {
	empty := func() genny.Generator[int] { return sequence.New[int]() }
	for name, g := range map[string]genny.Generator[int]{
		"allow":  constrain.AllowInterval(empty(), []int{1}),
		"block":  constrain.BlockInterval(empty(), []int{1}),
		"repeat": constrain.RemoveRepetitions(empty()),
		"dup":    constrain.Duplicates(empty(), 3),
		"shot":   constrain.Shotgun(empty(), 0.5),
	} {
		if !g.Done() {
			t.Errorf("%s: not done", name)
		}
		if v := g.Generate(); v != 0 {
			t.Errorf("%s: %d", name, v)
		}
	}
}

func TestRemoveRepetitions(t *testing.T) {
	genny.Seed(2)
	got := take(constrain.RemoveRepetitions(function.New(func() int { return genny.Intn(3) })), 500)
	for i := 1; i < len(got); i++ {
		if got[i] == got[i-1] {
			t.Fatalf("repeat at %d", i)
		}
	}
	// strings work too
	g := constrain.RemoveRepetitions(sequence.New("a", "a", "b"))
	if g.Generate() != "a" || g.Generate() != "b" {
		t.Error("strings")
	}
}

func TestDuplicates(t *testing.T) {
	genny.Seed(3)
	src := function.New(func() int { return genny.Intn(6) })
	got := take(constrain.Duplicates(src, 6), 120)
	for g := 0; g < len(got); g += 6 {
		grp := slices.Clone(got[g : g+6])
		slices.Sort(grp)
		if !slices.Equal(grp, []int{0, 1, 2, 3, 4, 5}) {
			t.Fatalf("group %d: %v", g/6, got[g:g+6])
		}
	}
	// group sizes that follow a generator: 1, 2, 3, 1, 2, 3
	sizes := sequence.NewLoop(1, 2, 3)
	src = function.New(func() int { return genny.Intn(3) })
	got = take(constrain.DuplicatesVar(src, sizes.Generate), 12)
	pos := 0
	for _, n := range []int{1, 2, 3, 1, 2} {
		grp := got[pos : pos+n]
		for i := range grp {
			if slices.Index(grp, grp[i]) != i {
				t.Fatalf("duplicate in %v (all %v)", grp, got)
			}
		}
		pos += n
	}
}

func TestDuplicatesInfeasible(t *testing.T) {
	// one value only, group of 3: the limit lets the repeats through
	got := take(constrain.Duplicates(function.New(func() int { return 4 }), 3, constrain.Stop(3)), 6)
	if !slices.Equal(got, []int{4, 4, 4, 4, 4, 4}) {
		t.Errorf("%v", got)
	}
}

func TestDuplicatesReset(t *testing.T) {
	g := constrain.Duplicates(sequence.NewLoop(1, 1, 2), 2, constrain.Stop(2))
	a := take(g, 4)
	g.Reset()
	b := take(g, 4)
	if !slices.Equal(a, b) {
		t.Errorf("%v %v", a, b)
	}
}

func TestShotgun(t *testing.T) {
	genny.Seed(4)
	// drops about half of 0..: the output is a subsequence in order, with holes
	g := constrain.Shotgun(sequence.NewLoop(0, 1, 2, 3, 4, 5, 6, 7, 8, 9), 0.5)
	got := take(g, 200)
	skipped := 0
	for i := 1; i < len(got); i++ {
		skipped += (got[i]-got[i-1]+10)%10 - 1
	}
	if frac := float64(skipped) / float64(skipped+len(got)); frac < 0.4 || frac > 0.6 {
		t.Errorf("dropped %.2f", frac)
	}

	// 0 drops nothing, > 1 and NaN are clamped
	if got := take(constrain.Shotgun(sequence.NewLoop(1, 2, 3), 0), 6); !slices.Equal(got, []int{1, 2, 3, 1, 2, 3}) {
		t.Errorf("%v", got)
	}
	if got := take(constrain.Shotgun(sequence.NewLoop(1, 2, 3), math.NaN()), 3); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("NaN: %v", got)
	}
	// everything dropped: gives up after the limit, still a value
	g = constrain.Shotgun(sequence.NewLoop(1, 2, 3), 7, constrain.Stop(4))
	if v := g.Generate(); v != 1+3%3 && v != 4%3 {
		t.Errorf("%d", v)
	}
}

func TestPanics(t *testing.T) {
	tests := map[string]func(){
		"duplicates size 0":   func() { constrain.Duplicates(sequence.New(1), 0) },
		"random lo > hi":      func() { constrain.RandomIntervals(5, 1, 1) },
		"random no intervals": func() { constrain.RandomIntervals(0, 5) },
		"random negative":     func() { constrain.RandomIntervals(0, 5, -1) },
	}
	for name, f := range tests {
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

func TestWrapperState(t *testing.T) {
	g := constrain.AllowInterval(sequence.NewLoop(60, 62, 64, 65), []int{2})
	if !g.Continuous() {
		t.Error("loop is continuous")
	}
	a := take(g, 5)
	g.Reset()
	if b := take(g, 5); !slices.Equal(a, b) {
		t.Errorf("%v %v", a, b)
	}
}
