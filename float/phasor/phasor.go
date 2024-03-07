package phasor

type Phasor struct {
	startPhase float64
	phase      float64
	inc        float64
	fs         float64
}

func New(fc float64, fs float64, phase float64) *Phasor {
	return &Phasor{
		startPhase: phase,
		phase:      phase,
		inc:        fc / fs,
	}
}

func (p *Phasor) SetFrequency(fc float64, fs float64) {
	p.inc = fc / fs
}

func (p *Phasor) SetPhase(phase float64) {
	p.startPhase = phase
	p.phase = phase
}

func (p *Phasor) PhaseOffset(offset float64) {
	p.phase += offset
	p.phaseWrap()
}

func (p *Phasor) phaseWrap() {
	for p.phase >= 1.0 {
		p.phase -= 1.0
	}

	for p.phase < 0.0 {
		p.phase += 1.0
	}
}

func (p *Phasor) Generate() (out float64) {
	out = p.phase

	p.phase += p.inc

	p.phaseWrap()

	return
}

func (p *Phasor) Continuous() bool {
	return true
}

func (p *Phasor) Done() bool {
	return false
}

func (p *Phasor) Reset() {
	p.phase = p.startPhase
}
