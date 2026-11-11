package pulse

import (
	"math"
	"slices"
	"testing"

	"github.com/almerlucke/genny/sequence"
)

func TestHierarchyKnown(t *testing.T) {
	for _, c := range []struct {
		div  []int
		want []int
	}{
		{nil, []int{0}},
		{[]int{1}, []int{0}},
		{[]int{2}, []int{1, 0}},
		{[]int{3}, []int{2, 0, 1}},
		{[]int{5}, []int{4, 0, 3, 1, 2}},
		{[]int{7}, []int{6, 0, 4, 2, 5, 1, 3}},
		{[]int{2, 2}, []int{3, 0, 2, 1}},
		{[]int{4}, []int{3, 0, 2, 1}},
		{[]int{2, 3}, []int{5, 0, 2, 4, 1, 3}},
		{[]int{1, 2, 1, 2}, []int{3, 0, 2, 1}},
	} {
		if got := Hierarchy(c.div...); !slices.Equal(got, c.want) {
			t.Errorf("Hierarchy(%v) = %v, want %v", c.div, got, c.want)
		}
	}
}

func TestHierarchyIsPermutation(t *testing.T) {
	for _, div := range [][]int{{3, 2, 2}, {4, 4}, {2, 2, 2, 2}, {5, 3}, {11}, {2, 3, 2}, {9}, {6, 2}} {
		ind := Hierarchy(div...)
		sorted := slices.Clone(ind)
		slices.Sort(sorted)
		for i, v := range sorted {
			if v != i {
				t.Fatalf("Hierarchy(%v) = %v is not a permutation", div, ind)
			}
		}
		if ind[0] != len(ind)-1 {
			t.Errorf("Hierarchy(%v): first pulse is %d", div, ind[0])
		}
	}
	if n := len(Hierarchy(3, 2, 2)); n != 12 {
		t.Errorf("3,2,2 has %d pulses", n)
	}
}

func TestHierarchyLevels(t *testing.T) {
	// 4/4 in sixteenths: beats before eighths before sixteenths
	ind := Hierarchy(2, 2, 2, 2)
	for i, v := range ind {
		for j, w := range ind {
			if level(i) < level(j) && v <= w {
				t.Errorf("pulse %d (level %d) is not more important than %d (level %d)", i, level(i), j, level(j))
			}
		}
	}
}

// level of a position in 16 steps: 0 for the first pulse, 4 for the odd pulses
func level(n int) int {
	if n == 0 {
		return 0
	}
	c := 0
	for n%2 == 0 {
		n /= 2
		c++
	}
	return 4 - c
}

func TestHierarchyPanics(t *testing.T) {
	for _, div := range [][]int{{0}, {3, -1}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%v did not panic", div)
				}
			}()
			Hierarchy(div...)
		}()
	}
}

func TestThin(t *testing.T) {
	ind := Hierarchy(2, 2)
	for _, c := range []struct {
		level int
		want  []bool
	}{
		{-5, []bool{true, true, true, true}},
		{0, []bool{true, true, true, true}},
		{1, []bool{true, false, true, true}},
		{2, []bool{true, false, true, false}},
		{3, []bool{true, false, false, false}},
		{4, []bool{false, false, false, false}},
		{100, []bool{false, false, false, false}},
	} {
		if got := Thin(ind, c.level); !slices.Equal(got, c.want) {
			t.Errorf("Thin level %d = %v, want %v", c.level, got, c.want)
		}
	}
	if len(Thin(nil, 1)) != 0 {
		t.Error("thin of nothing")
	}
}

func TestVelocity(t *testing.T) {
	got := Velocity(Hierarchy(2, 2), 0.2, 0.8)
	want := []float64{0.8, 0.2, 0.6, 0.4}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Fatalf("Velocity = %v, want %v", got, want)
		}
	}
	if got := Velocity([]int{0}, 0.2, 0.8); got[0] != 0.8 {
		t.Errorf("single pulse = %v", got)
	}
	if len(Velocity(nil, 0, 1)) != 0 {
		t.Error("empty")
	}
}

func TestInterpolate(t *testing.T) {
	a, b := Hierarchy(2, 2), []int{0, 3, 1, 2}
	for _, c := range []struct {
		rate float64
		want []int
	}{
		{0, a}, {1, b}, {-3, a}, {7, b}, {0.5, []int{2, 2, 2, 2}},
	} {
		if got := Interpolate(a, b, c.rate); !slices.Equal(got, c.want) {
			t.Errorf("rate %v = %v, want %v", c.rate, got, c.want)
		}
	}
	defer func() {
		if recover() == nil {
			t.Error("length mismatch did not panic")
		}
	}()
	Interpolate(a, []int{1}, 0.5)
}

func take(g *Gate, n int) []bool {
	var out []bool
	for range n {
		out = append(out, g.Generate())
	}
	return out
}

func TestConstGate(t *testing.T) {
	ind := Hierarchy(2, 2)
	got := take(NewConstGate(ind, 2), 8)
	want := []bool{true, false, true, false, true, false, true, false}
	if !slices.Equal(got, want) {
		t.Errorf("gate = %v", got)
	}
	// beyond the range
	if g := take(NewConstGate(ind, 99), 4); slices.Contains(g, true) {
		t.Errorf("level 99 = %v", g)
	}
	if g := take(NewConstGate(ind, -4), 4); slices.Contains(g, false) {
		t.Errorf("level -4 = %v", g)
	}
}

func TestGateFollowsGenerator(t *testing.T) {
	ind := Hierarchy(2, 2)
	// per pulse: level 4 (nothing), 0 (all), 3, 1.5 (rounds to 2)
	g := NewGate(ind, sequence.NewLoop(4.0, 0, 3, 1.5))
	got := take(g, 4)
	want := []bool{false, true, false, false}
	if !slices.Equal(got, want) {
		t.Errorf("gate = %v, want %v", got, want)
	}
	g.Reset()
	if got := take(g, 4); !slices.Equal(got, want) {
		t.Errorf("after reset = %v", got)
	}
	if !g.Continuous() || g.Done() {
		t.Error("gate loops forever")
	}
}

func TestEmptyGate(t *testing.T) {
	g := NewConstGate(nil, 1)
	if !g.Done() || g.Continuous() || g.Generate() {
		t.Error("empty gate should be done and false")
	}
	g.Reset()
	if !g.Done() {
		t.Error("done after reset")
	}
}

func TestGateNaNLevel(t *testing.T) {
	g := NewGate(Hierarchy(2), sequence.NewLoop(math.NaN()))
	if got := take(g, 2); !slices.Equal(got, []bool{true, true}) {
		t.Errorf("nan level = %v", got)
	}
}
