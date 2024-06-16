package interp

import (
	"github.com/almerlucke/genny/float"
)

type Method int

const (
	Linear Method = iota
	Cubic
)

type interpolationHistory [4]float64

func (h interpolationHistory) interpolateLinear(t float64) float64 {
	return h[0] + (h[1]-h[0])*t
}

func (h interpolationHistory) interpolateCubic(t float64) float64 {
	t2 := t * t
	t3 := t * t2
	return (2*t3-3*t2+1)*h[1] + (t3-2*t2+t)*h[0] + (-2*t3+3*t2)*h[2] + (t3-t2)*h[3]
}

type Interpolator struct {
	generator float.FrameGenerator
	dt        float64
	t         float64
	method    Method
	history   []interpolationHistory
	outVector []float64
	done      bool
}

func New(generator float.FrameGenerator, method Method, dt float64) *Interpolator {
	numDimensions := generator.Dimensions()

	interpol := &Interpolator{
		generator: generator,
		dt:        dt,
		method:    method,
		history:   make([]interpolationHistory, numDimensions),
		outVector: make([]float64, numDimensions),
	}

	for dim := 0; dim < numDimensions; dim++ {
		interpol.history[dim] = interpolationHistory{}
	}

	interpol.initialize()

	return interpol
}

func (ipol *Interpolator) Dimensions() int {
	return ipol.generator.Dimensions()
}

func (ipol *Interpolator) SetDelta(dt float64) {
	ipol.dt = dt
}

func (ipol *Interpolator) initialize() {
	switch ipol.method {
	case Linear:
		for dim, v := range ipol.generator.Generate() {
			ipol.history[dim][0] = v
		}
		for dim, v := range ipol.generator.Generate() {
			ipol.history[dim][1] = v
		}
	case Cubic:
		for dim, v := range ipol.generator.Generate() {
			ipol.history[dim][0] = v
			ipol.history[dim][1] = v
		}
		for dim, v := range ipol.generator.Generate() {
			ipol.history[dim][2] = v
		}
		for dim, v := range ipol.generator.Generate() {
			ipol.history[dim][3] = v
		}
	}
}

func (ipol *Interpolator) updateHistory() {
	switch ipol.method {
	case Linear:
		for dim, v := range ipol.generator.Generate() {
			ipol.history[dim][0] = ipol.history[dim][1]
			ipol.history[dim][1] = v
		}
	case Cubic:
		for dim, v := range ipol.generator.Generate() {
			ipol.history[dim][0] = ipol.history[dim][1]
			ipol.history[dim][1] = ipol.history[dim][2]
			ipol.history[dim][2] = ipol.history[dim][3]
			ipol.history[dim][3] = v
		}
	}
}

func (ipol *Interpolator) interpolate(t float64) []float64 {
	switch ipol.method {
	case Linear:
		for dim := 0; dim < ipol.Dimensions(); dim++ {
			ipol.outVector[dim] = ipol.history[dim].interpolateLinear(t)
		}
	case Cubic:
		for dim := 0; dim < ipol.Dimensions(); dim++ {
			ipol.outVector[dim] = ipol.history[dim].interpolateCubic(t)
		}
	}

	return ipol.outVector
}

func (ipol *Interpolator) Generate() []float64 {
	out := ipol.interpolate(ipol.t)

	ipol.t += ipol.dt

	if ipol.t >= 1.0 {
		ipol.t -= 1.0

		if !ipol.generator.Continuous() {
			ipol.done = ipol.generator.Done()
		}

		ipol.updateHistory()
	}

	return out
}

func (ipol *Interpolator) Continuous() bool {
	return ipol.generator.Continuous()
}

func (ipol *Interpolator) Done() bool {
	return ipol.done
}

func (ipol *Interpolator) Reset() {
	ipol.generator.Reset()
	ipol.t = 0.0
	ipol.done = false

	for dim := 0; dim < ipol.Dimensions(); dim++ {
		ipol.history[dim] = interpolationHistory{}
	}

	ipol.initialize()
}
