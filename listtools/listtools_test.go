package listtools

import (
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/sequence"
)

func takeN[T any](g genny.Generator[T], n int) []T {
	var out []T
	for range n {
		out = append(out, g.Generate())
	}
	return out
}

func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s did not panic", name)
		}
	}()
	f()
}

func TestJump(t *testing.T) {
	v := []int{0, 1, 2, 3, 4, 5, 6, 7}
	for _, c := range []struct {
		chunk, step, leap int
		want              []int
	}{
		{3, 1, 4, []int{0, 1, 2, 4, 5, 6, 0, 1, 2}},
		{2, 2, 1, []int{0, 2, 1, 3, 2, 4}},
		{2, 1, -3, []int{0, 1, 5, 6, 2, 3}},   // negative leap wraps
		{3, -1, 0, []int{0, 7, 6, 0, 7, 6}},   // negative step wraps
		{1, 1, 3, []int{0, 3, 6, 1, 4, 7, 2}}, // chunk 1 is a stride
	} {
		got := takeN[int](NewJump(v, c.chunk, c.step, c.leap), len(c.want))
		if !slices.Equal(got, c.want) {
			t.Errorf("Jump(%d,%d,%d) = %v, want %v", c.chunk, c.step, c.leap, got, c.want)
		}
	}
}

func TestJumpEdges(t *testing.T) {
	g := NewJump([]string{}, 2, 1, 1)
	if !g.Done() || g.Continuous() || g.Generate() != "" {
		t.Error("empty jump should be done and generate the zero value")
	}
	mustPanic(t, "chunk 0", func() { NewJump([]int{1}, 0, 1, 1) })

	j := NewJump([]int{1, 2, 3}, 2, 1, 1)
	first := takeN[int](j, 5)
	j.Reset()
	if !slices.Equal(takeN[int](j, 5), first) {
		t.Error("reset")
	}
	if j.Done() || !j.Continuous() {
		t.Error("jump loops")
	}
}

func TestStutter(t *testing.T) {
	g := NewStutter[int](sequence.NewLoop(1, 2, 3, 4), 2, 2)
	got := takeN[int](g, 10)
	want := []int{1, 2, 2, 2, 3, 4, 4, 4, 1, 2}
	if !slices.Equal(got, want) {
		t.Errorf("stutter = %v, want %v", got, want)
	}
	// k = 0 changes nothing
	g = NewStutter[int](sequence.NewLoop(1, 2, 3), 1, 0)
	if got := takeN[int](g, 3); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("k=0 = %v", got)
	}
	mustPanic(t, "n 0", func() { NewStutter[int](sequence.NewLoop(1), 0, 1) })
	mustPanic(t, "k < 0", func() { NewStutter[int](sequence.NewLoop(1), 1, -1) })
}

func TestStutterDone(t *testing.T) {
	g := NewStutter[int](sequence.New(1, 2), 2, 2)
	var got []int
	for !g.Done() && len(got) < 20 {
		got = append(got, g.Generate())
	}
	if !slices.Equal(got, []int{1, 2, 2, 2}) {
		t.Errorf("stutter at the end = %v", got)
	}
	g.Reset()
	if g.Done() || g.Generate() != 1 {
		t.Error("reset")
	}
	e := NewStutter[int](sequence.New[int](), 2, 2)
	if !e.Done() || e.Generate() != 0 {
		t.Error("empty")
	}
}

func TestGroup(t *testing.T) {
	g := NewGroupN[string](sequence.NewLoop("a", "b"), 3)
	if got := takeN[string](g, 8); !slices.Equal(got, []string{"a", "a", "a", "b", "b", "b", "a", "a"}) {
		t.Errorf("group = %v", got)
	}
	// k from a generator, and values below 1 count as 1
	g = NewGroup[string](sequence.NewLoop("a", "b", "c"), sequence.NewLoop(1, 2, 0, -4))
	if got := takeN[string](g, 6); !slices.Equal(got, []string{"a", "b", "b", "c", "a", "b"}) {
		t.Errorf("group gen = %v", got)
	}
	once := NewGroupN[int](sequence.New(7, 8), 2)
	var got []int
	for !once.Done() {
		got = append(got, once.Generate())
	}
	if !slices.Equal(got, []int{7, 7, 8, 8}) {
		t.Errorf("group once = %v", got)
	}
	once.Reset()
	if once.Done() {
		t.Error("reset")
	}
}

func TestGrow(t *testing.T) {
	for _, c := range []struct {
		name           string
		start          float64
		mode           Mode
		amount, lo, hi float64
		limit          Limit
		want           []float64
	}{
		{"geometric clip", 1, Geometric, 2, 0, 10, Clip, []float64{1, 2, 4, 8, 10, 10}},
		{"geometric restart", 1, Geometric, 2, 0, 10, Restart, []float64{1, 2, 4, 8, 1, 2}},
		{"arithmetic clip", 2, Arithmetic, 3, 0, 10, Clip, []float64{2, 5, 8, 10, 10}},
		{"arithmetic down", 5, Arithmetic, -2, 0, 10, Clip, []float64{5, 3, 1, 0, 0}},
		{"arithmetic restart", 5, Arithmetic, -2, 0, 10, Restart, []float64{5, 3, 1, 5, 3}},
		{"start clamped", 50, Arithmetic, 1, 0, 10, Clip, []float64{10, 10}},
		{"shrinking", 8, Geometric, 0.5, 1, 10, Clip, []float64{8, 4, 2, 1, 1}},
	} {
		g := NewGrow(c.start, c.mode, c.amount, c.lo, c.hi, c.limit)
		if got := takeN[float64](g, len(c.want)); !slices.Equal(got, c.want) {
			t.Errorf("%s = %v, want %v", c.name, got, c.want)
		}
	}
	g := NewGrow(1, Arithmetic, 1, 0, 10, Clip)
	takeN[float64](g, 4)
	g.Reset()
	if g.Generate() != 1 {
		t.Error("reset")
	}
	mustPanic(t, "lo > hi", func() { NewGrow(1, Geometric, 2, 5, 1, Clip) })
	// infinity or NaN steps do not escape the range
	h := NewGrow(1, Geometric, math.Inf(1), 0, 10, Clip)
	for _, v := range takeN[float64](h, 4) {
		if v < 0 || v > 10 || math.IsNaN(v) {
			t.Errorf("out of range %v", v)
		}
	}
	n := NewGrow(0, Geometric, math.Inf(1), 0, 10, Clip) // 0 * Inf = NaN
	for _, v := range takeN[float64](n, 3) {
		if math.IsNaN(v) {
			t.Error("NaN")
		}
	}
}

func TestGatherUntil(t *testing.T) {
	for _, c := range []struct {
		name  string
		gen   genny.Generator[float64]
		total float64
		want  []float64
	}{
		{"trim", sequence.NewLoop(1.5, 1), 4, []float64{1.5, 1, 1.5}},
		{"trim last", sequence.NewLoop(1.5, 1), 3.5, []float64{1.5, 1, 1}},
		{"exact", sequence.NewLoop(1.0, 2), 3, []float64{1, 2}},
		{"rests", sequence.NewLoop(-1.5, 1), 3, []float64{-1.5, 1, -0.5}},
		{"trim a rest", sequence.NewLoop(-3.0), 2, []float64{-2}},
		{"generator ends", sequence.New(1.0, 1), 5, []float64{1, 1}},
		{"zero total", sequence.NewLoop(1.0), 0, nil},
		{"negative total", sequence.NewLoop(1.0), -2, nil},
		{"empty", sequence.New[float64](), 4, nil},
		{"zeros", constant.New(0.0), 4, nil},
	} {
		got := GatherUntil(c.gen, c.total)
		if !slices.Equal(got, c.want) {
			t.Errorf("%s = %v, want %v", c.name, got, c.want)
		}
	}
	sum := 0.0
	for _, v := range GatherUntil(sequence.NewLoop(0.3, 0.7, 0.45), 8) {
		sum += math.Abs(v)
	}
	if math.Abs(sum-8) > 1e-9 {
		t.Errorf("sum = %v", sum)
	}
}

func TestPlusMin(t *testing.T) {
	genny.Seed(2)
	all := NewPlusMin(sequence.NewLoop(1.0, 2, 3), 1)
	for _, v := range takeN[float64](all, 6) {
		if v >= 0 {
			t.Errorf("p=1 gave %v", v)
		}
	}
	none := NewPlusMin(sequence.NewLoop(1.0, -2), 0)
	if got := takeN[float64](none, 4); !slices.Equal(got, []float64{1, -2, 1, -2}) {
		t.Errorf("p=0 = %v", got)
	}
	// half of the values are rests, roughly
	g := NewPlusMin(constant.New(1.0), 0.5)
	neg := 0
	for _, v := range takeN[float64](g, 1000) {
		if v < 0 {
			neg++
		}
	}
	if neg < 400 || neg > 600 {
		t.Errorf("%d of 1000 are rests", neg)
	}
	if g.Done() || !g.Continuous() {
		t.Error("delegates")
	}
	over := NewPlusMin(constant.New(2.0), 7) // clamped
	if over.Generate() != -2 {
		t.Error("p > 1")
	}
}

func TestPositiveNegative(t *testing.T) {
	g := NewPositiveNegative(sequence.NewLoop(1.0, -2, 3), 2, 1)
	if got := takeN[float64](g, 7); !slices.Equal(got, []float64{1, 2, -3, 1, 2, -3, 1}) {
		t.Errorf("= %v", got)
	}
	g.Reset()
	if g.Generate() != 1 {
		t.Error("reset")
	}
	allNeg := NewPositiveNegative(constant.New(2.0), 0, 2)
	if got := takeN[float64](allNeg, 2); !slices.Equal(got, []float64{-2, -2}) {
		t.Errorf("= %v", got)
	}
	allPos := NewPositiveNegative(constant.New(-2.0), 1, 0)
	if got := takeN[float64](allPos, 2); !slices.Equal(got, []float64{2, 2}) {
		t.Errorf("= %v", got)
	}
	mustPanic(t, "0,0", func() { NewPositiveNegative(constant.New(1.0), 0, 0) })
	mustPanic(t, "negative", func() { NewPositiveNegative(constant.New(1.0), -1, 2) })
}
