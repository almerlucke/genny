package tests

import (
	"fmt"
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/and"
	"github.com/almerlucke/genny/arpeggio"
	"github.com/almerlucke/genny/bucket"
	"github.com/almerlucke/genny/combine"
	"github.com/almerlucke/genny/continuous"
	"github.com/almerlucke/genny/conv"
	"github.com/almerlucke/genny/flatten"
	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/genny/float/export/plot"
	"github.com/almerlucke/genny/float/interp"
	"github.com/almerlucke/genny/float/iter"
	"github.com/almerlucke/genny/float/iter/updaters/chaos"
	"github.com/almerlucke/genny/float/phasor"
	"github.com/almerlucke/genny/float/ramp"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/markov"
	"github.com/almerlucke/genny/or"
	"github.com/almerlucke/genny/repeat"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/transform"
	"github.com/almerlucke/genny/unwrap"
	"github.com/almerlucke/genny/walk"
	"gonum.org/v1/plot/vg"
	"log"
	"math/rand"
	"testing"
)

type StringCaster struct{}

func (sc *StringCaster) Convert(f float64) string {
	return fmt.Sprintf("test cast: %f", f)
}

func TestGens(t *testing.T) {
	var g genny.Generator[float64] = sequence.New(1.0, 2.0, 3.0)
	for !g.Done() {
		log.Printf("sequence: %f", g.Generate())
	}

	g = function.New(nil, func(_ any) float64 { return rand.Float64() })
	for i := 0; i < 10; i++ {
		log.Printf("function: %f", g.Generate())
	}

	g = repeat.NewWithFunc[float64](sequence.New(2.0, 3.0, 4.0), func() int { return 3 + rand.Intn(5) })
	for !g.Done() {
		log.Printf("repeat: %f", g.Generate())
	}

	g = and.New[float64](
		repeat.New[float64](sequence.New(2.0, 3.0, 4.0), 4),
		repeat.New[float64](sequence.New(8.0, 9.0, 10.0), 4),
	)
	for !g.Done() {
		log.Printf("and: %f", g.Generate())
	}

	g = bucket.New(bucket.Random, 2.0, 3.0, 4.0, 5.0)
	for i := 0; i < 10; i++ {
		log.Printf("bucket random: %f", g.Generate())
	}

	g = bucket.NewContinuous(bucket.Indexed, 2.0, 3.0, 4.0, 5.0)
	for i := 0; i < 10; i++ {
		log.Printf("bucket indexed: %f", g.Generate())
	}

	g = or.New[float64](or.Indexed, sequence.New(2.0, 3.0, 4.0), sequence.New(12.0, 13.0, 14.0))
	for !g.Done() {
		log.Printf("or indexed: %f", g.Generate())
	}

	g = or.NewContinuous[float64](or.Indexed, sequence.NewContinuous(2.0, 3.0, 4.0), sequence.NewContinuous(12.0, 13.0, 14.0))
	for i := 0; i < 20; i++ {
		log.Printf("or indexed continuous: %f", g.Generate())
	}

	g = flatten.NewFlatten[float64](sequence.New[[]float64]([]float64{1.0, 2.0, 3.0}, []float64{4.0, 5.0, 6.0}))
	for !g.Done() {
		log.Printf("flatten: %f", g.Generate())
	}

	g = walk.New[float64](walk.NewMatrix(
		[]int{3, 3},
		[]float64{
			1.0, 2.0, 3.0,
			4.0, 5.0, 6.0,
			7.0, 8.0, 9.0,
		}))
	for i := 0; i < 10; i++ {
		log.Printf("walk: %f", g.Generate())
	}

	var f transform.Function[float64] = func(v float64) float64 { return v + 12 }
	g = transform.New[float64](sequence.New(2.0, 3.0, 4.0), f)
	for !g.Done() {
		log.Printf("transform: %f", g.Generate())
	}

	state1 := markov.NewProbabilityState(1.0)
	state2 := markov.NewProbabilityState(2.0)
	state3 := markov.NewProbabilityState(3.0)
	state4 := markov.NewProbabilityState(4.0)
	state1.SetProbabilities(state1, 0.1, state2, 0.9)
	state2.SetProbabilities(state2, 0.1, state3, 0.9)
	state3.SetProbabilities(state3, 0.1, state4, 0.9)
	state4.SetProbabilities(state4, 0.1, nil, 0.9)
	g = markov.New[float64](state1, 1)
	for !g.Done() {
		log.Printf("markov: %f", g.Generate())
	}

	g = arpeggio.New([]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}, arpeggio.Converge, arpeggio.Exclusive, false)
	for !g.Done() {
		log.Printf("arpeggio: %f", g.Generate())
	}

	gsc := conv.New[float64, string](ramp.New(10, 0.0, 1.0, 2.0), &StringCaster{})
	for !gsc.Done() {
		log.Printf("ramp + cast: %s", gsc.Generate())
	}

	g = continuous.New[float64](sequence.New(1.0, 2.0, 3.0, 4.0))
	for i := 0; i < 10; i++ {
		log.Printf("continuous: %f", g.Generate())
	}

	gcm := combine.New[float64](sequence.New(1.0, 2.0, 3.0, 4.0), sequence.New(1.0, 2.0, 3.0, 4.0))
	for !gcm.Done() {
		log.Printf("combine: %v", gcm.Generate())
	}

	gg := unwrap.New[float64](sequence.New[genny.Generator[float64]](sequence.New(1.0, 2.0, 3.0), sequence.New(1.0, 2.0, 3.0)))
	for !gg.Done() {
		log.Printf("unwrap: %v", gg.Generate())
	}

	it := iter.New([]float64{0.1231}, chaos.NewVerhulstWithFunc(3.6951, chaos.Iter1))
	ipol := interp.New(it, interp.Linear, 0.1)
	for i := 0; i < 100; i++ {
		log.Printf("interpolator: %f", ipol.Generate()[0])
	}

	err := plot.Plot(
		float.ToFrame(phasor.New(1000.0, 44100.0, 0.0)),
		100,
		vg.Centimeter*10, vg.Centimeter*5,
		"points.png",
	)
	if err != nil {
		t.Errorf("plot error: %v", err)
	}
}
