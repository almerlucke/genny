package chaos_test

import (
	"math"
	"testing"

	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
)

func TestDuffingBounded(t *testing.T) {
	for i, f := range run([]float64{1, 0}, chaos.NewDuffing(), 20000) {
		if math.IsNaN(f[0]) || math.Abs(f[0]) > 3 || math.Abs(f[1]) > 3 {
			t.Fatalf("step %d: %v", i, f)
		}
	}
}

func TestDuffingUndrivenDecays(t *testing.T) {
	d := chaos.NewDuffing()
	d.Gamma = 0
	got := run([]float64{1.2, 0}, d, 4000)
	last := got[len(got)-1]
	// settles in one of the wells at x = +-1
	if math.Abs(math.Abs(last[0])-1) > 1e-3 || math.Abs(last[1]) > 1e-3 {
		t.Errorf("did not settle in a well: %v", last)
	}
}

func TestPendulum(t *testing.T) {
	// no damping, no drive, small swing: energy is kept by RK4 and the period is about 2 pi
	p := &chaos.Pendulum{Dt: 0.01}
	got := run([]float64{0.1, 0}, p, 629)
	if last := got[len(got)-1]; math.Abs(last[0]-0.1) > 1e-3 {
		t.Errorf("after one period: %v", last)
	}
	// driven: the wrapped angle stays in [-pi, pi)
	for i, f := range run([]float64{0.2, 0}, chaos.NewPendulum(), 20000) {
		if f[0] < -math.Pi || f[0] >= math.Pi || math.IsNaN(f[1]) {
			t.Fatalf("step %d: %v", i, f)
		}
	}
}

func TestStandardMap(t *testing.T) {
	// K = 0: theta moves by the constant p, p stays
	got := run([]float64{1, 0.5}, chaos.NewStandardMap(0), 3)
	if got[1][0] != 1.5 || got[1][1] != 0.5 || math.Abs(got[2][0]-2) > 1e-12 {
		t.Errorf("%v", got)
	}
	// always in [0, 2 pi)
	for i, f := range run([]float64{1, 0.5}, chaos.NewStandardMap(2), 5000) {
		for _, x := range f {
			if x < 0 || x >= 2*math.Pi {
				t.Fatalf("step %d: %v", i, f)
			}
		}
	}
}

func TestShortStateDoesNotPanic(t *testing.T) {
	for _, u := range []iter.Updater{chaos.NewDuffing(), chaos.NewPendulum(), chaos.NewStandardMap(1)} {
		u.Update([]float64{1})
		u.Update(nil)
	}
}

func TestProject(t *testing.T) {
	tests := []struct {
		axes chaos.Axes
		dims int
		want []float64
	}{
		{chaos.AxisX, 1, []float64{1}},
		{chaos.AxisY, 1, []float64{0.5}},
		{chaos.AxisBoth, 2, []float64{1, 0.5}},
	}
	for _, tt := range tests {
		p := chaos.Project(iter.New([]float64{1, 0.5}, chaos.NewStandardMap(0)), tt.axes)
		got := p.Generate()
		if p.Dimensions() != tt.dims || len(got) != tt.dims {
			t.Fatalf("axes %v: dims %d, %v", tt.axes, p.Dimensions(), got)
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("axes %v: %v want %v", tt.axes, got, tt.want)
			}
		}
	}
	// one value only: the missing y is zero
	p := chaos.Project(iter.New([]float64{3}, chaos.NewStandardMap(0)), chaos.AxisY)
	if got := p.Generate(); got[0] != 0 {
		t.Errorf("%v", got)
	}
}
