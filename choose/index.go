package choose

import (
	"math"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/dist"
)

// Distribution decides which part of a list is favoured. Given the length n of the list it returns a
// position in [0, 1) (0 is the start of the list, 1 the end) that Index turns into an element.
// Values outside [0, 1) are clamped. The constructors below make the usual ones, you can also write
// your own function
type Distribution func(n int) float64

// Uniform picks every element with the same chance
func Uniform() Distribution {
	return func(int) float64 { return genny.Float64() }
}

// Linear favours the start of the list: the chance falls linearly to zero at the end
func Linear() Distribution {
	return func(int) float64 { return min(genny.Float64(), genny.Float64()) }
}

// Exponential favours the start of the list, with a chance that falls exponentially. A larger rate
// falls faster, near 0 is almost uniform. It panics if rate is not positive
func Exponential(rate float64) Distribution {
	if !(rate > 0) {
		panic("choose: rate must be positive")
	}
	top := 1 - math.Exp(-rate)
	return func(int) float64 { return -math.Log(1-genny.Float64()*top) / rate }
}

// Gaussian favours elements around index mean, with a standard deviation of sd in index units (sd 1
// is mostly the elements next to the mean). Draws outside the list are drawn again a few times,
// after that they are put at the nearest end. It panics if sd is negative
func Gaussian(mean, sd float64) Distribution {
	if sd < 0 {
		panic("choose: sd must not be negative")
	}
	return func(n int) float64 {
		x := 0.0
		for range 16 {
			u1, u2 := 1-genny.Float64(), genny.Float64()
			x = mean + sd*math.Sqrt(-2*math.Log(u1))*math.Cos(2*math.Pi*u2)
			// index i covers [i-0.5, i+0.5)
			if x >= -0.5 && x < float64(n)-0.5 {
				break
			}
		}
		return (min(max(x, -0.5), float64(n)-0.5) + 0.5) / float64(n)
	}
}

// Beta favours the ends (a and b below 1), the middle (above 1) or one side (a smaller than b is the
// start), see dist.SampleBeta. It panics if a or b is not positive
func Beta(a, b float64) Distribution {
	if !(a > 0) || !(b > 0) {
		panic("choose: beta parameters must be positive")
	}
	return func(int) float64 { return dist.SampleBeta(a, b) }
}

// Cauchy favours the position center (in [0, 1]) with half of the picks within scale of it, and
// now and then a far pick. Draws outside the list are drawn again a few times, after that they are
// put at the nearest end. It panics if scale is not positive
func Cauchy(center, scale float64) Distribution {
	g := dist.NewCauchy(center, scale).Retry(0, math.Nextafter(1, 0), 16)
	return func(int) float64 { return g.Generate() }
}

// Weibull favours the start of the list for a shape below 1, and a position around spread (in [0, 1])
// for a larger shape. Draws past the end are drawn again a few times, after that they are put
// at the end. It panics if spread or shape is not positive
func Weibull(spread, shape float64) Distribution {
	g := dist.NewWeibull(0, spread, shape).Retry(0, math.Nextafter(1, 0), 16)
	return func(int) float64 { return g.Generate() }
}

// Triangle favours position mode (in [0, 1]): the chance rises linearly from the start to mode and falls
// again to the end. It panics if mode is outside [0, 1]
func Triangle(mode float64) Distribution {
	if !(mode >= 0 && mode <= 1) {
		panic("choose: mode must be in [0, 1]")
	}
	return func(int) float64 {
		u := genny.Float64()
		if u < mode {
			return math.Sqrt(u * mode)
		}
		return 1 - math.Sqrt((1-u)*(1-mode))
	}
}

// Ratio picks element i with a chance proportional to the integer weight i: Ratio(4, 2, 1) picks the
// first element four times as often as the third. Elements beyond the weights are never picked. It
// panics if a weight is negative or all are zero
func Ratio(weights ...int) Distribution {
	total := 0
	for _, w := range weights {
		if w < 0 {
			panic("choose: weights must not be negative")
		}
		total += w
	}
	if total == 0 {
		panic("choose: at least one weight must be positive")
	}
	return func(n int) float64 {
		r := genny.Intn(total)
		for i, w := range weights {
			if r < w {
				// pick the middle of the element's slot, whatever the length of the list
				return (float64(min(i, n-1)) + 0.5) / float64(n)
			}
			r -= w
		}
		return 0
	}
}

// Index picks an element of a list by position, with a Distribution that favours some part of the
// list. An empty list is done at once and generates the zero value
type Index[T any] struct {
	items []T
	d     Distribution
}

// NewIndex creates an Index over a copy of items. It panics if d is nil
func NewIndex[T any](items []T, d Distribution) *Index[T] {
	if d == nil {
		panic("choose: distribution must not be nil")
	}
	return &Index[T]{items: append([]T(nil), items...), d: d}
}

// Pick returns the index of the next pick, or -1 for an empty list
func (x *Index[T]) Pick() int {
	n := len(x.items)
	if n == 0 {
		return -1
	}
	p := x.d(n)
	if math.IsNaN(p) {
		p = 0
	}
	return min(max(int(math.Floor(p*float64(n))), 0), n-1)
}

func (x *Index[T]) Generate() (v T) {
	if i := x.Pick(); i >= 0 {
		return x.items[i]
	}
	return
}

func (x *Index[T]) Continuous() bool { return len(x.items) > 0 }

func (x *Index[T]) Done() bool { return len(x.items) == 0 }

func (x *Index[T]) Reset() {}

// ChangingWeights picks elements of a list with weights, and changes the weights every N picks: it
// cycles through a list of weight sets, so the preference moves through the list over time. An empty
// list is done at once and generates the zero value
type ChangingWeights[T any] struct {
	items []T
	every int
	sets  [][]float64
	count int
}

// NewChangingWeights creates a choice over a copy of items that uses the first weight set for the
// first every picks, the second for the next every picks, and so on, around. A set that is shorter than
// the list gives the missing elements weight 0. It panics if every is less than 1, there is no set,
// a weight is negative or NaN, or a set has no positive weight
func NewChangingWeights[T any](items []T, every int, sets ...[]float64) *ChangingWeights[T] {
	if every < 1 {
		panic("choose: every must be at least 1")
	}
	if len(sets) == 0 {
		panic("choose: at least one weight set is needed")
	}
	c := &ChangingWeights[T]{items: append([]T(nil), items...), every: every}
	for _, s := range sets {
		total := 0.0
		for _, w := range s {
			if !(w >= 0) {
				panic("choose: weights must not be negative")
			}
			total += w
		}
		if !(total > 0) {
			panic("choose: a weight set needs a positive weight")
		}
		c.sets = append(c.sets, append([]float64(nil), s...))
	}
	return c
}

func (c *ChangingWeights[T]) Generate() (v T) {
	if len(c.items) == 0 {
		return
	}
	set := c.sets[(c.count/c.every)%len(c.sets)]
	c.count++
	n := min(len(set), len(c.items))
	total := 0.0
	for _, w := range set[:n] {
		total += w
	}
	if total == 0 { // the set only has weights past the end of the list
		return c.items[0]
	}
	r := genny.Float64() * total
	pick := n - 1
	for i, w := range set[:n] {
		if w > 0 && r < w {
			pick = i
			break
		}
		r -= w
	}
	// the last positive weight when rounding left a remainder
	for pick > 0 && set[pick] == 0 {
		pick--
	}
	return c.items[pick]
}

func (c *ChangingWeights[T]) Continuous() bool { return len(c.items) > 0 }

func (c *ChangingWeights[T]) Done() bool { return len(c.items) == 0 }

// Reset goes back to the first weight set
func (c *ChangingWeights[T]) Reset() { c.count = 0 }
