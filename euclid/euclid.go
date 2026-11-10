// Package euclid generates Euclidean rhythms: pulses spread as evenly as possible over steps
package euclid

// Pattern returns the Euclidean rhythm (Bjorklund) for the given number of pulses spread over steps.
// The pattern is rotated left by rotation steps (negative rotates right). Pulses are clamped to [0, steps].
// It panics if steps is less than 1
func Pattern(steps, pulses, rotation int) []bool {
	if steps < 1 {
		panic("euclid: steps must be at least 1")
	}
	pulses = min(max(pulses, 0), steps)

	a := make([][]bool, pulses)
	for i := range a {
		a[i] = []bool{true}
	}
	b := make([][]bool, steps-pulses)
	for i := range b {
		b[i] = []bool{false}
	}

	for len(b) > 1 && len(a) > 0 {
		n := min(len(a), len(b))
		merged := make([][]bool, n)
		for i := range n {
			merged[i] = append(append([]bool{}, a[i]...), b[i]...)
		}
		var rest [][]bool
		if len(a) > n {
			rest = a[n:]
		} else {
			rest = b[n:]
		}
		a, b = merged, rest
	}

	flat := make([]bool, 0, steps)
	for _, g := range a {
		flat = append(flat, g...)
	}
	for _, g := range b {
		flat = append(flat, g...)
	}

	out := make([]bool, steps)
	for i := range out {
		out[((i-rotation)%steps+steps)%steps] = flat[i]
	}
	return out
}

// Euclid is a looping generator of the steps of a Euclidean rhythm, true is a pulse
type Euclid struct {
	pattern []bool
	index   int
}

// New creates a looping Euclidean rhythm generator, see Pattern
func New(steps, pulses, rotation int) *Euclid {
	return &Euclid{pattern: Pattern(steps, pulses, rotation)}
}

// SetPattern changes the rhythm, the position is kept (wrapped) so the change is heard at the next step
func (e *Euclid) SetPattern(steps, pulses, rotation int) {
	e.pattern = Pattern(steps, pulses, rotation)
	e.index %= len(e.pattern)
}

// Pattern returns a copy of the current pattern
func (e *Euclid) Pattern() []bool {
	return append([]bool{}, e.pattern...)
}

func (e *Euclid) Generate() bool {
	v := e.pattern[e.index]
	e.index = (e.index + 1) % len(e.pattern)
	return v
}

func (e *Euclid) Continuous() bool { return true }

func (e *Euclid) Done() bool { return false }

func (e *Euclid) Reset() { e.index = 0 }
