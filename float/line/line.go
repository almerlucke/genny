package line

type Line struct {
	from      float64
	to        float64
	cur       float64
	stepsLeft int
	numSteps  int
	delta     float64
}

func New(start float64) *Line {
	return &Line{
		from: start,
		cur:  start,
		to:   start,
	}
}

func (l *Line) Generate() float64 {
	out := l.cur

	if l.stepsLeft == 0 {
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
	return l.stepsLeft == 0
}

func (l *Line) Reset() {
	l.cur = l.from
	l.stepsLeft = l.numSteps
}

func (l *Line) From(from float64) {
	l.from = from
	l.cur = from
	l.to = from
}

func (l *Line) To(to float64, numSteps int) {
	l.from = l.to
	l.to = to
	l.cur = l.from
	l.delta = (l.to - l.from) / float64(numSteps)
	l.stepsLeft = numSteps
}
