package chaos_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
)

func run(init []float64, u iter.Updater, n int) [][]float64 {
	it := iter.New(init, u)
	var out [][]float64
	for i := 0; i < n; i++ {
		out = append(out, append([]float64(nil), it.Generate()...))
	}
	return out
}

func TestVerhulst(t *testing.T) {
	// x' = 2x(1-x)
	got := run([]float64{0.25}, chaos.NewVerhulst(2), 4)
	want := [][]float64{{0.25}, {0.375}, {0.46875}, {2 * 0.46875 * (1 - 0.46875)}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	// fixed point at 1 - 1/a
	got = run([]float64{0.5}, chaos.NewVerhulst(2), 3)
	if !reflect.DeepEqual(got, [][]float64{{0.5}, {0.5}, {0.5}}) {
		t.Fatalf("got %v", got)
	}
	// logistic map at a=4 stays in [0,1]
	for _, v := range run([]float64{0.3}, chaos.NewVerhulst(4), 500) {
		if v[0] < 0 || v[0] > 1 {
			t.Fatalf("%v left [0,1]", v)
		}
	}
}

func TestVerhulstFunctions(t *testing.T) {
	tests := []struct {
		name string
		f    func(float64) float64
		x    float64
		want float64
	}{
		{"1", chaos.Iter1, 0.5, 0.25},
		{"1a", chaos.Iter1a, 0.5, 0.375},
		{"1b", chaos.Iter1b, math.Pi / 2, math.Pi / 2},
		{"1c", chaos.Iter1c, 0, 0},
		{"1c at 1", chaos.Iter1c, 1, math.E},
		{"1d", chaos.Iter1d, 1, 0.25},
		{"1e", chaos.Iter1e, 0.75, 0.75 * 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f(tt.x); math.Abs(got-tt.want) > 1e-12 {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
	got := run([]float64{3}, chaos.NewVerhulstWithFunc(0.5, func(x float64) float64 { return x + 1 }), 3)
	if !reflect.DeepEqual(got, [][]float64{{3}, {2}, {1.5}}) {
		t.Fatalf("got %v", got)
	}
}

func TestHenon(t *testing.T) {
	// x' = y + 1 - a*x^2, y' = b*x
	got := run([]float64{0, 0}, chaos.NewHenon(1.4, 0.3), 4)
	want := [][]float64{{0, 0}, {1, 0}, {-0.4, 0.3}, {0.3 + 1 - 1.4*0.4*0.4, -0.12}}
	for i := range want {
		for d := range want[i] {
			if math.Abs(got[i][d]-want[i][d]) > 1e-12 {
				t.Fatalf("got %v want %v", got, want)
			}
		}
	}
	// the classic attractor stays bounded
	for _, v := range run([]float64{0.1, 0.1}, chaos.NewHenon(1.4, 0.3), 2000) {
		if math.Abs(v[0]) > 2 || math.Abs(v[1]) > 1 {
			t.Fatalf("%v left the attractor", v)
		}
	}
	// custom function replaces x*x
	got = run([]float64{2, 1}, chaos.NewHenonWithFunc(1, 2, func(x float64) float64 { return x }), 2)
	if !reflect.DeepEqual(got, [][]float64{{2, 1}, {0, 4}}) {
		t.Fatalf("got %v", got)
	}
}

func TestAronson(t *testing.T) {
	// x' = y, y' = a*y*(1-f(x)) with f the identity
	got := run([]float64{2, 3}, chaos.NewAronson(1), 3)
	want := [][]float64{{2, 3}, {3, -3}, {-3, 6}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	got = run([]float64{1, 2}, chaos.NewAronsonWithFunc(0.5, func(x float64) float64 { return 2 * x }), 2)
	if !reflect.DeepEqual(got, [][]float64{{1, 2}, {2, -1}}) {
		t.Fatalf("got %v", got)
	}
}
