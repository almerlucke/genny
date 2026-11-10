// Package line provides Line, a generator that moves linearly from one value to the next in a given
// number of steps.
package line

// Line moves from the current value to a target in a number of steps, then holds the target
type Line struct {
	from      float64
	to        float64
	cur       float64
	stepsLeft int
	numSteps  int
	delta     float64
}

// New creates a line that stays at start until To is called
func New(start float64) *Line {
	return &Line{
		from: start,
		cur:  start,
		to:   start,
	}
}

func (l *Line) Generate() float64 {
	out := l.cur

	if l.stepsLeft <= 0 {
		goto returnValue
	}

	l.cur += l.delta
	l.stepsLeft--

returnValue:
	return out
}

func (l *Line) Continuous() bool {
	return false
}

func (l *Line) Done() bool {
	return l.stepsLeft <= 0
}

func (l *Line) Reset() {
	l.cur = l.from
	l.stepsLeft = l.numSteps
}

// From jumps to a value and stays there
func (l *Line) From(from float64) {
	l.from = from
	l.cur = from
	l.to = from
	l.numSteps = 0
	l.stepsLeft = 0
}

// To starts a line from the last target to a new target in numSteps steps
func (l *Line) To(to float64, numSteps int) {
	l.from = l.to
	l.to = to
	l.cur = l.from
	l.delta = (l.to - l.from) / float64(numSteps)
	l.stepsLeft = numSteps
	l.numSteps = numSteps
}
