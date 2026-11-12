package automaton_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/automaton"
	"github.com/almerlucke/genny/scale"
)

func TestParseRule(t *testing.T) {
	for _, c := range []struct {
		in, out string
		ok      bool
	}{
		{"B3/S23", "B3/S23", true},
		{"b36/s23", "B36/S23", true},
		{" B2/S ", "B2/S", true},
		{"B/S", "B/S", true},
		{"B3S23", "", false},
		{"S23/B3", "", false},
		{"B9/S2", "", false},
		{"B3/S2x", "", false},
		{"", "", false},
	} {
		r, err := automaton.ParseRule(c.in)
		if (err == nil) != c.ok {
			t.Errorf("ParseRule(%q) error = %v", c.in, err)
			continue
		}
		if c.ok && r.String() != c.out {
			t.Errorf("ParseRule(%q) = %s, want %s", c.in, r, c.out)
		}
	}
}

func TestNewLifePanics(t *testing.T) {
	for name, f := range map[string]func(){
		"no width":  func() { automaton.NewLife(0, 4, automaton.Conway) },
		"no height": func() { automaton.NewLife(4, 0, automaton.Conway) },
		"bad rule":  func() { automaton.NewLife(4, 4, "life") },
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

func TestBlinkerHasPeriodTwo(t *testing.T) {
	l := automaton.NewLife(7, 7, automaton.Conway)
	l.Place(automaton.Blinker, 2, 3)
	start := l.String()
	l.Step()
	if l.String() == start {
		t.Fatal("blinker did not change")
	}
	if got := strings.Count(l.String(), "#"); got != 3 {
		t.Fatalf("blinker has %d cells", got)
	}
	l.Step()
	if l.String() != start {
		t.Errorf("blinker after 2 generations:\n%s\nwant:\n%s", l, start)
	}
}

func TestStillLifeAndToad(t *testing.T) {
	l := automaton.NewLife(8, 8, automaton.Conway)
	l.Place(automaton.Block, 3, 3)
	start := l.String()
	for range 5 {
		l.Step()
	}
	if l.String() != start {
		t.Error("block moved")
	}
	l = automaton.NewLife(8, 8, automaton.Conway)
	l.Place(automaton.Toad, 2, 3)
	start = l.String()
	l.Step()
	l.Step()
	if l.String() != start {
		t.Error("toad is not period 2")
	}
}

func TestGliderMovesDiagonally(t *testing.T) {
	const size = 10
	l := automaton.NewLife(size, size, automaton.Conway)
	l.Place(automaton.Glider, 1, 1)
	shifted := func(dx int) string {
		g := automaton.NewLife(size, size, automaton.Conway)
		g.Place(automaton.Glider, 1+dx, 1+dx)
		return g.String()
	}
	for gen := 1; gen <= 4*size; gen++ { // a full lap around the torus
		l.Step()
		if gen%4 == 0 {
			if l.String() != shifted(gen/4) {
				t.Fatalf("generation %d:\n%s\nwant:\n%s", gen, l, shifted(gen/4))
			}
		}
		if l.Population() != 5 {
			t.Fatalf("generation %d has %d cells", gen, l.Population())
		}
	}
	if l.String() != shifted(0) {
		t.Error("the glider did not come back after a lap")
	}
}

func TestLWSS(t *testing.T) {
	l := automaton.NewLife(14, 10, automaton.Conway)
	l.Place(automaton.LWSS, 1, 2)
	for range 4 {
		l.Step()
	}
	want := automaton.NewLife(14, 10, automaton.Conway)
	want.Place(automaton.LWSS, 3, 2)
	if l.String() != want.String() {
		t.Errorf("after 4 generations:\n%s\nwant:\n%s", l, want)
	}
}

func TestRPentominoGrows(t *testing.T) {
	l := automaton.NewLife(40, 40, automaton.Conway)
	l.Place(automaton.RPentomino, 19, 19)
	var pops []int
	for range 10 {
		l.Step()
		pops = append(pops, l.Population())
	}
	// the known start of the population of the R-pentomino: 5, 6, 7, 9, 8, 9, 12, 11, 18, ...
	if want := []int{6, 7, 9, 8, 9, 12, 11, 18}; !slices.Equal(pops[:8], want) {
		t.Errorf("population %v, want %v", pops, want)
	}
}

func TestHighLifeReplicatorDiffersFromLife(t *testing.T) {
	// B6 only exists in HighLife: a dead cell with six neighbours is born
	hl := automaton.NewLife(7, 7, automaton.HighLife)
	cv := automaton.NewLife(7, 7, automaton.Conway)
	for _, l := range []*automaton.Life{hl, cv} {
		l.Place(automaton.Pattern{"###", "#.#", ".#."}, 2, 2)
	}
	hl.Step()
	cv.Step()
	if hl.String() == cv.String() {
		t.Error("HighLife and Life gave the same generation")
	}
}

func TestSeedsRule(t *testing.T) {
	l := automaton.NewLife(6, 6, automaton.Seeds)
	l.Place(automaton.Pattern{"##"}, 2, 2)
	l.Step()
	// every live cell dies, the cells above and below both are born (each has two neighbours)
	if got := l.Population(); got != 4 {
		t.Errorf("population %d, want 4:\n%s", got, l)
	}
}

func TestLifeWrapsAndResets(t *testing.T) {
	l := automaton.NewLife(5, 4, automaton.Conway)
	l.Set(-1, -1, true)
	if !l.Alive(4, 3) || !l.Alive(9, 7) {
		t.Error("coordinates do not wrap")
	}
	l.Place(automaton.Blinker, 0, 0)
	start := l.String()
	l.Step()
	l.Step()
	l.Step()
	l.Reset()
	if l.String() != start || l.Generation() != 0 {
		t.Error("Reset did not restore the start")
	}
	if l.Width() != 5 || l.Height() != 4 {
		t.Error("size")
	}
}

func TestRowAndColumn(t *testing.T) {
	l := automaton.NewLife(4, 3, automaton.Conway)
	l.Set(1, 0, true)
	l.Set(1, 2, true)
	l.Set(3, 2, true)
	if got := l.Row(2); !slices.Equal(got, []bool{false, true, false, true}) {
		t.Errorf("row %v", got)
	}
	if got := l.Column(1); !slices.Equal(got, []bool{true, false, true}) {
		t.Errorf("column %v", got)
	}
	if got := l.Row(-1); !slices.Equal(got, l.Row(2)) {
		t.Error("row does not wrap")
	}
}

func TestLifeRandomDensity(t *testing.T) {
	genny.Seed(8)
	for _, c := range []struct {
		d        float64
		min, max int
	}{{0, 0, 0}, {-1, 0, 0}, {1, 400, 400}, {3, 400, 400}, {0.25, 70, 130}} {
		l := automaton.NewLifeRandom(20, 20, automaton.Conway, c.d)
		if p := l.Population(); p < c.min || p > c.max {
			t.Errorf("density %v gave %d cells", c.d, p)
		}
	}
}

func TestLifeGateReadsRowThenSteps(t *testing.T) {
	l := automaton.NewLife(5, 5, automaton.Conway)
	l.Place(automaton.Blinker, 1, 2) // row 2: .###.
	g := automaton.NewRowGate(l, 2)
	var got []bool
	for range 10 {
		got = append(got, g.Generate())
	}
	// first cycle is the row of generation 0, the second one of generation 1 (the blinker is vertical, so
	// only the middle cell of row 2 is alive)
	want := []bool{false, true, true, true, false, false, false, true, false, false}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if l.Generation() != 2 {
		t.Errorf("generation %d, want 2", l.Generation())
	}
	g.Reset()
	if l.Generation() != 0 {
		t.Error("Reset did not reset the grid")
	}
	if g.Generate() { // first cell of the row again
		t.Error("gate did not restart")
	}
	if !g.Generate() { // second cell
		t.Error("gate did not restart")
	}
}

func TestColumnGateLength(t *testing.T) {
	l := automaton.NewLife(3, 6, automaton.Conway)
	g := automaton.NewColumnGate(l, 1)
	for range 12 {
		g.Generate()
	}
	if l.Generation() != 2 {
		t.Errorf("a column of 6 played twice should give 2 generations, got %d", l.Generation())
	}
	if !g.Continuous() || g.Done() {
		t.Error("gate should be endless")
	}
}

func TestLifePitches(t *testing.T) {
	key := scale.New(60, scale.Major)
	l := automaton.NewLife(6, 4, automaton.Conway)
	l.Set(0, 1, true)
	l.Set(2, 1, true)
	l.Set(5, 1, true)
	p := automaton.NewLifePitches(l, 1, key, 0, 2, 4)
	var got []float64
	var rests []bool
	for range 6 {
		o := p.Generate()
		got = append(got, o.Value)
		rests = append(rests, o.Rest)
	}
	wantRests := []bool{false, true, false, true, true, false}
	if !slices.Equal(rests, wantRests) {
		t.Errorf("rests %v, want %v", rests, wantRests)
	}
	// column 0 -> degree 0, column 2 -> degrees[2] = 4 (G), column 5 -> degrees[5%3=2] = 4
	if got[0] != 60 || got[2] != 67 || got[5] != 67 {
		t.Errorf("pitches %v", got)
	}

	plain := automaton.NewLifePitches(automaton.NewLifeRandom(4, 2, automaton.Conway, 1), 0, key)
	if o := plain.Generate(); o.Rest || o.Value != 60 {
		t.Errorf("first cell without degrees: %+v", o)
	}
	if o := plain.Generate(); o.Value != 62 {
		t.Errorf("second cell without degrees: %+v", o)
	}
}

func TestLifePitchesNilScalePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("no panic")
		}
	}()
	automaton.NewLifePitches(automaton.NewLife(3, 3, automaton.Conway), 0, nil)
}

func TestLifeDensity(t *testing.T) {
	l := automaton.NewLife(4, 4, automaton.Conway)
	l.Place(automaton.Block, 0, 0)
	d := automaton.NewLifeDensity(l)
	if v := d.Generate(); v != 0.25 {
		t.Errorf("density %v, want 0.25", v)
	}
	if l.Generation() != 1 {
		t.Error("density did not step")
	}
	d.Reset()
	if l.Generation() != 0 {
		t.Error("Reset")
	}
}

func TestLifeSeeded(t *testing.T) {
	run := func(seed uint64) string {
		genny.Seed(seed)
		l := automaton.NewLifeRandom(12, 12, automaton.Conway, 0.4)
		for range 5 {
			l.Step()
		}
		return l.String()
	}
	if run(1) != run(1) || run(1) == run(2) {
		t.Error("random life is not seeded")
	}
}
