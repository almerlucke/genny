package walk_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/walk"
)

func TestMatrixLayout(t *testing.T) {
	m := walk.NewMatrix([]int{3, 2}, []int{0, 1, 2, 3, 4, 5})
	if !reflect.DeepEqual(m.Dimensions(), []int{3, 2}) {
		t.Fatal("dimensions")
	}
	// the first dimension varies fastest
	for j := 0; j < 2; j++ {
		for i := 0; i < 3; i++ {
			if got := m.ValueAtIndex([]int{i, j}); got != i+3*j {
				t.Fatalf("[%d,%d] = %d", i, j, got)
			}
		}
	}
	m3 := walk.NewMatrix([]int{2, 2, 2}, []int{0, 1, 2, 3, 4, 5, 6, 7})
	if m3.ValueAtIndex([]int{1, 0, 1}) != 5 || m3.ValueAtIndex([]int{0, 1, 1}) != 6 {
		t.Fatal("3 dimensional layout")
	}
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// identity matrix: the value at (x, y) is x + w*y
func TestWalkMovesOneStepInsideTheContainer(t *testing.T) {
	genny.Seed(1)
	const w, h = 4, 3
	vals := make([]int, w*h)
	for i := range vals {
		vals[i] = i
	}
	wk := walk.New[int](walk.NewMatrix([]int{w, h}, vals))
	if !wk.Continuous() || wk.Done() {
		t.Fatal("walk is continuous and never done")
	}
	wk.Reset()

	px, py := -1, -1
	moved := map[string]bool{}
	for i := 0; i < 2000; i++ {
		v := wk.Generate()
		if v < 0 || v >= w*h {
			t.Fatalf("value %d outside the container", v)
		}
		x, y := v%w, v/w
		if px >= 0 {
			if abs(x-px)+abs(y-py) != 1 {
				t.Fatalf("step from (%d,%d) to (%d,%d) is not one step", px, py, x, y)
			}
			if x != px {
				moved["x"] = true
			} else {
				moved["y"] = true
			}
		}
		px, py = x, y
	}
	if !moved["x"] || !moved["y"] {
		t.Fatalf("expected moves in both dimensions: %v", moved)
	}
}

func TestWalkOneDimensionVisitsEverything(t *testing.T) {
	genny.Seed(2)
	wk := walk.New[int](walk.NewMatrix([]int{5}, []int{0, 1, 2, 3, 4}))
	seen := map[int]bool{}
	prev := -1
	for i := 0; i < 500; i++ {
		v := wk.Generate()
		if prev >= 0 && abs(v-prev) != 1 {
			t.Fatalf("%d -> %d", prev, v)
		}
		prev = v
		seen[v] = true
	}
	if len(seen) != 5 {
		t.Fatalf("seen %v", seen)
	}
}

func TestWalkSizeTwoDimension(t *testing.T) {
	genny.Seed(3)
	wk := walk.New[int](walk.NewMatrix([]int{2}, []int{10, 20}))
	prev := wk.Generate()
	for i := 0; i < 20; i++ {
		v := wk.Generate()
		if v == prev || (v != 10 && v != 20) {
			t.Fatalf("%d -> %d", prev, v)
		}
		prev = v
	}
}

// a dimension of size 1 has no room to move, the walk must stay inside the container
func TestWalkSizeOneDimensions(t *testing.T) {
	genny.Seed(4)
	tests := []struct {
		name string
		dims []int
		n    int
	}{
		{"single cell", []int{1}, 1},
		{"1x1", []int{1, 1}, 1},
		{"1x4", []int{1, 4}, 4},
		{"4x1", []int{4, 1}, 4},
		{"3x1x3", []int{3, 1, 3}, 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vals := make([]int, tt.n)
			for i := range vals {
				vals[i] = i
			}
			wk := walk.New[int](walk.NewMatrix(tt.dims, vals))
			for i := 0; i < 500; i++ {
				v := wk.Generate()
				if v < 0 || v >= tt.n {
					t.Fatalf("value %d outside the container", v)
				}
			}
		})
	}
}

func TestWalkNoDimensions(t *testing.T) {
	wk := walk.New[int](walk.NewMatrix(nil, []int{42}))
	if wk.Generate() != 42 {
		t.Fatal("a zero dimensional container has one cell")
	}
}

func TestWalkSeedRepeatable(t *testing.T) {
	run := func() []int {
		genny.Seed(21)
		vals := make([]int, 16)
		for i := range vals {
			vals[i] = i
		}
		wk := walk.New[int](walk.NewMatrix([]int{4, 4}, vals))
		out := make([]int, 50)
		for i := range out {
			out[i] = wk.Generate()
		}
		return out
	}
	if a, b := run(), run(); !reflect.DeepEqual(a, b) {
		t.Fatal("same seed must give the same walk")
	}
}
