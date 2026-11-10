package shape_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/sequence"
)

type shaperFunc func(float64) float64

func (f shaperFunc) Shape(x float64) float64 { return f(x) }

type pairs struct {
	frames [][]float64
	i      int
	resets int
}

func (p *pairs) Generate() []float64 { f := p.frames[p.i%len(p.frames)]; p.i++; return f }
func (p *pairs) Dimensions() int     { return 2 }
func (p *pairs) Continuous() bool    { return false }
func (p *pairs) Done() bool          { return p.i >= len(p.frames) }
func (p *pairs) Reset()              { p.i = 0; p.resets++ }

func TestShapeShapesEveryDimension(t *testing.T) {
	src := &pairs{frames: [][]float64{{1, 2}, {3, 4}}}
	var _ float.FrameGenerator = shape.New(src, shaperFunc(nil))
	s := shape.New(src, shaperFunc(func(x float64) float64 { return x * 10 }))
	if s.Dimensions() != 2 || s.Continuous() || s.Done() {
		t.Fatal("dimensions / state follow the source")
	}
	if got := s.Generate(); !reflect.DeepEqual(got, []float64{10, 20}) {
		t.Fatalf("got %v", got)
	}
	if got := s.Generate(); !reflect.DeepEqual(got, []float64{30, 40}) {
		t.Fatalf("got %v", got)
	}
	if !s.Done() {
		t.Fatal("done follows the source")
	}
	s.Reset()
	if src.resets != 1 || s.Done() {
		t.Fatal("reset follows the source")
	}
}

func TestSingle(t *testing.T) {
	s := shape.NewSingle(sequence.New(1.0, 2.0, 3.0), shaperFunc(func(x float64) float64 { return x*x + 1 }))
	if s.Continuous() || s.Done() {
		t.Fatal("state follows the source")
	}
	for _, want := range []float64{2, 5, 10} {
		if got := s.Generate(); got != want {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	if !s.Done() {
		t.Fatal("done")
	}
	s.Reset()
	if s.Done() || s.Generate() != 2 {
		t.Fatal("reset")
	}
}
