package main

import (
	"math"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/choose"
	"github.com/almerlucke/genny/float/dist"
	"github.com/almerlucke/genny/float/tendency"
)

// tpHistogram returns a function that walks over the bins of a normalized histogram (the highest bin
// is 1) of n draws, one bin per call, so a histogram can be drawn with lines
func tpHistogram(draw func() float64, lo, hi float64, bins, n int) func() float64 {
	h := make([]float64, bins)
	for range n {
		v := draw()
		if v < lo || v >= hi {
			continue
		}
		h[int((v-lo)/(hi-lo)*float64(bins))]++
	}
	peak := 0.0
	for _, c := range h {
		peak = max(peak, c)
	}
	i := -1
	return func() float64 {
		i = (i + 1) % bins
		return h[i] / peak
	}
}

// plotsTendency writes the plots of the tendency masks, the beta, Cauchy and Weibull distributions,
// pink noise and the choice by index to dir
func plotsTendency(dir string) {
	genny.Seed(11)

	// a mask: the band between its bounds and the values drawn in it, uniform and beta (0.4, 0.4)
	mask := tendency.NewMask(false,
		tendency.Segment{N: 60, A1: 50, A2: 50, Z1: 20, Z2: 80},
		tendency.Segment{N: 60, A1: 20, A2: 80, Z1: 45, Z2: 55},
		tendency.Segment{N: 60, A1: 45, A2: 55, Z1: 70, Z2: 90},
	)
	i, j := -1, -1
	lowBound := func() float64 { i++; lo, _ := mask.At(i); return lo }
	highBound := func() float64 { j++; _, hi := mask.At(j); return hi }
	lines(dir, "tendency_mask.png", "Tendency mask: bounds and values", mask.Len(),
		[]string{"lower bound", "upper bound", "uniform", "beta 0.4, 0.4"},
		lowBound, highBound, tendency.NewValue(mask).Generate, tendency.NewBetaValue(mask, 0.4, 0.4).Generate)

	// the same mask for a list of 12 elements (percentages), with an inner mask kept free
	outer := tendency.NewMask(true, tendency.Segment{N: 120, A1: 0, A2: 100, Z1: 0, Z2: 100})
	inner := tendency.NewMask(true, tendency.Segment{N: 120, A1: 40, A2: 60, Z1: 10, Z2: 90})
	k := -1
	lines(dir, "tendency_two_masks.png", "Two masks: the outer band without the inner band", 120,
		[]string{"values", "inner low", "inner high"},
		tendency.NewTwoMasks(outer, inner, 20).Generate,
		func() float64 { k++; lo, _ := inner.At(k); return lo },
		func() float64 { _, hi := inner.At(k); return hi })

	// anchor: a shape stretched over 40 values that starts and ends at given values
	shape := []float64{0, 1, 0.3, 0.8, 0}
	anchored := tendency.Anchor(shape, 40, 20, 70, 0, 100)
	sprayed := tendency.Spray(anchored, 6)
	a, s := -1, -1
	lines(dir, "tendency_anchor.png", "Anchor and spray", len(anchored), []string{"anchor", "spray"},
		func() float64 { a++; return anchored[a] }, func() float64 { s++; return sprayed[s] })

	// histograms of the distributions
	const bins, n = 40, 200000
	genny.Seed(12)
	var betas []func() float64
	for _, p := range [][2]float64{{1, 1}, {2, 5}, {5, 2}, {0.4, 0.4}} {
		betas = append(betas, tpHistogram(dist.NewBeta(p[0], p[1]).Generate, 0, 1, bins, n))
	}
	lines(dir, "dist_beta.png", "Beta distribution (histogram)", bins,
		[]string{"1, 1", "2, 5", "5, 2", "0.4, 0.4"}, betas...)

	var cauchys []func() float64
	for _, sc := range []float64{0.5, 1, 2} {
		cauchys = append(cauchys, tpHistogram(dist.NewCauchy(0, sc).Generate, -10, 10, bins, n))
	}
	lines(dir, "dist_cauchy.png", "Cauchy distribution, location 0 (histogram, -10 to 10)", bins,
		[]string{"scale 0.5", "scale 1", "scale 2"}, cauchys...)

	var weibulls []func() float64
	for _, sh := range []float64{0.7, 1, 2, 4} {
		weibulls = append(weibulls, tpHistogram(dist.NewWeibull(0, 2, sh).Generate, 0, 6, bins, n))
	}
	lines(dir, "dist_weibull.png", "Weibull distribution, spread 2 (histogram, 0 to 6)", bins,
		[]string{"shape 0.7", "shape 1", "shape 2", "shape 4"}, weibulls...)

	// pink noise against white noise, and the Lehmer generator
	genny.Seed(13)
	lines(dir, "pink.png", "Pink (1/f) noise", 400, []string{"pink 8 rows", "pink 3 rows", "white"},
		dist.NewPink(8).Generate, dist.NewPink(3).Generate, genny.Float64)
	lehmer := dist.NewLehmer(7)
	lines(dir, "lehmer.png", "Lehmer generator", 200, []string{"lehmer"}, lehmer.Generate)

	// how often each of 16 elements is picked by index
	items := make([]int, 16)
	for i := range items {
		items[i] = i
	}
	pick := func(d choose.Distribution) func() float64 {
		x := choose.NewIndex(items, d)
		return tpHistogram(func() float64 { return float64(x.Generate()) + 0.5 }, 0, 16, 16, 50000)
	}
	genny.Seed(14)
	lines(dir, "choose_index.png", "Choice by index (16 elements)", 16,
		[]string{"linear", "exponential 3", "gaussian 8, 2", "triangle 0.7", "ratio 4:2:1", "weibull 0.4, 2"},
		pick(choose.Linear()), pick(choose.Exponential(3)), pick(choose.Gaussian(8, 2)),
		pick(choose.Triangle(0.7)), pick(choose.Ratio(4, 2, 1)), pick(choose.Weibull(0.4, 2)))

	// changing weights: the picked index over time, the preference moves every 40 picks
	cw := choose.NewChangingWeights(items, 40,
		[]float64{1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]float64{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0},
		[]float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1})
	lines(dir, "choose_changing.png", "Changing weights: the weights change every 40 picks", 240,
		[]string{"index"}, func() float64 { return math.Floor(float64(cw.Generate())) })
}
