package evolve

import (
	"math"

	"github.com/almerlucke/genny/scale"
)

// Smoothness rates how small the leaps between neighbouring notes are: 1 for a phrase on one note (or with
// fewer than 2 notes), falling towards 0 as the average leap grows (1 / (1 + mean leap))
func Smoothness() Fitness {
	return func(p []int) float64 {
		if len(p) < 2 {
			return 1
		}
		sum := 0
		for i := 1; i < len(p); i++ {
			sum += abs(p[i] - p[i-1])
		}
		return 1 / (1 + float64(sum)/float64(len(p)-1))
	}
}

// Arch is a contour that rises to the middle and falls again, n values in [0, 1], for Contour. A value
// of n below 1 gives no values
func Arch(n int) []float64 {
	out := make([]float64, max(n, 0))
	for i := range out {
		if n > 1 {
			out[i] = math.Sin(math.Pi * float64(i) / float64(n-1))
		} else {
			out[i] = 1
		}
	}
	return out
}

// Contour rates how well the shape of a phrase matches a target contour. Both are scaled to [0, 1]
// (lowest to highest note) and compared note by note, the target is stretched to the length of the phrase
// by interpolation. The result is 1 minus the mean difference. A flat phrase has no shape and scores 0
// unless the target is flat as well; an empty target or phrase scores 1
func Contour(target []float64) Fitness {
	t := normalize(target)
	return func(p []int) float64 {
		if len(p) == 0 || len(t) == 0 {
			return 1
		}
		f := make([]float64, len(p))
		for i, v := range p {
			f[i] = float64(v)
		}
		f = normalize(f)
		sum := 0.0
		for i, v := range f {
			pos := 0.0
			if len(p) > 1 {
				pos = float64(i) / float64(len(p)-1) * float64(len(t)-1)
			}
			lo := int(pos)
			hi := min(lo+1, len(t)-1)
			want := t[lo] + (t[hi]-t[lo])*(pos-float64(lo))
			sum += math.Abs(v - want)
		}
		return 1 - sum/float64(len(f))
	}
}

// normalize scales values to [0, 1]; equal values stay 0.5 for a flat contour
func normalize(v []float64) []float64 {
	out := make([]float64, len(v))
	if len(v) == 0 {
		return out
	}
	lo, hi := v[0], v[0]
	for _, x := range v {
		lo, hi = math.Min(lo, x), math.Max(hi, x)
	}
	for i, x := range v {
		if hi > lo {
			out[i] = (x - lo) / (hi - lo)
		} else {
			out[i] = 0.5
		}
	}
	return out
}

// InScale rates the share of notes that are notes of the scale, for phrases of pitches in semitones: 1 when
// all are in the scale. An empty phrase scores 1
func InScale(s *scale.Scale) Fitness {
	return func(p []int) float64 {
		if len(p) == 0 {
			return 1
		}
		n := 0
		for _, v := range p {
			if s.Quantize(float64(v)) == float64(v) {
				n++
			}
		}
		return float64(n) / float64(len(p))
	}
}

// IntervalVariety rates how many different leaps (ignoring direction) the phrase uses, as a share of the
// number of leaps: a phrase in which no interval occurs twice scores 1. A phrase with fewer than 2 notes
// scores 1
func IntervalVariety() Fitness {
	return func(p []int) float64 {
		if len(p) < 2 {
			return 1
		}
		seen := map[int]bool{}
		for i := 1; i < len(p); i++ {
			seen[abs(p[i]-p[i-1])] = true
		}
		return float64(len(seen)) / float64(len(p)-1)
	}
}

// Target rates the closeness to a given phrase: 1 for the phrase itself, 1 / (1 + mean distance) in
// general. Notes of the phrase beyond the target (or missing from it) count as a distance of 1 each
func Target(phrase []int) Fitness {
	want := append([]int{}, phrase...)
	return func(p []int) float64 {
		n := max(len(p), len(want))
		if n == 0 {
			return 1
		}
		sum := 0
		for i := range n {
			if i < len(p) && i < len(want) {
				sum += abs(p[i] - want[i])
			} else {
				sum++
			}
		}
		return 1 / (1 + float64(sum)/float64(n))
	}
}

// Part is a fitness function with a weight, see Weighted
type Part struct {
	Fitness Fitness
	Weight  float64
}

// Weight pairs a fitness function with a weight, to pass to Weighted
func Weight(f Fitness, weight float64) Part { return Part{f, weight} }

// Weighted combines fitness functions into their weighted mean: with parts in [0, 1] the result is in
// [0, 1] too. It panics if a weight is negative or NaN or a fitness function is nil; no parts, or only
// weights of 0, give a fitness of 0
func Weighted(parts ...Part) Fitness {
	total := 0.0
	for _, p := range parts {
		if p.Fitness == nil || !(p.Weight >= 0) {
			panic("evolve: weighted parts need a fitness function and a weight of 0 or more")
		}
		total += p.Weight
	}
	parts = append([]Part{}, parts...)
	return func(phrase []int) float64 {
		if total == 0 {
			return 0
		}
		sum := 0.0
		for _, p := range parts {
			if p.Weight > 0 {
				sum += p.Weight * p.Fitness(phrase)
			}
		}
		return sum / total
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
