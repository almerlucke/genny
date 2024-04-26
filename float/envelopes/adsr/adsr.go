package adsr

import "math"

type stage int

const (
	attack stage = iota
	decay
	sustain
	release
	idle
)

type ReleaseMode int

const (
	Automatic ReleaseMode = iota // automatic release when sustain duration has passed
	Duration                     // explicit release when duration has passed
	NoteOff                      // release when Release() is called (note off)
)

const (
	shapeMult = 2.0
)

type step struct {
	level    float64
	duration float64 // duration in milliseconds
	shape    float64
}

type Setting struct {
	AttackLevel     float64
	AttackDuration  float64
	AttackShape     float64
	DecayLevel      float64
	DecayDuration   float64
	DecayShape      float64
	SkipDecay       bool
	SustainDuration float64
	SkipSustain     bool
	ReleaseDuration float64
	ReleaseShape    float64
}

func NewSetting(attackLevel float64, attackDuration float64, decayLevel float64, decayDuration float64, sustainDuration float64, releaseDuration float64) *Setting {
	return &Setting{
		AttackLevel:     attackLevel,
		AttackDuration:  attackDuration,
		AttackShape:     0.0,
		DecayLevel:      decayLevel,
		DecayDuration:   decayDuration,
		DecayShape:      0.0,
		SustainDuration: sustainDuration,
		ReleaseDuration: releaseDuration,
		ReleaseShape:    0.0,
	}
}

func NewARSetting(attackLevel float64, attackDuration float64, sustainDuration float64, releaseDuration float64) *Setting {
	return &Setting{
		AttackLevel:     attackLevel,
		AttackDuration:  attackDuration,
		AttackShape:     0.0,
		SkipDecay:       true,
		SustainDuration: sustainDuration,
		ReleaseDuration: releaseDuration,
		ReleaseShape:    0.0,
	}
}

func NewADSetting(attackLevel float64, attackDuration float64, releaseDuration float64) *Setting {
	return &Setting{
		AttackLevel:     attackLevel,
		AttackDuration:  attackDuration,
		AttackShape:     0.0,
		SkipDecay:       true,
		SkipSustain:     true,
		ReleaseDuration: releaseDuration,
		ReleaseShape:    0.0,
	}
}

func (setting *Setting) stepForStage(currentStage stage) (step, stage) {
	var stageStep step

	switch currentStage {
	case attack:
		stageStep = step{
			level: setting.AttackLevel, duration: setting.AttackDuration, shape: setting.AttackShape,
		}
	case decay:
		if !setting.SkipDecay {
			stageStep = step{
				level: setting.DecayLevel, duration: setting.DecayDuration, shape: setting.DecayShape,
			}
			break
		}
		currentStage++
		fallthrough
	case sustain:
		if !setting.SkipSustain {
			level := setting.DecayLevel
			if setting.SkipDecay {
				level = setting.AttackLevel
			}
			stageStep = step{
				level: level, duration: setting.SustainDuration, shape: 0.0,
			}
			break
		}
		currentStage++
		fallthrough
	case release:
		stageStep = step{
			level: 0.0, duration: setting.ReleaseDuration, shape: setting.ReleaseShape,
		}
	default:
		break
	}

	return stageStep, currentStage
}

type ADSR struct {
	setting     *Setting
	releaseMode ReleaseMode
	stage       stage
	releaseCnt  int64
	sr          float64
	duration    float64
	level       float64
	stepCnt     int64
	exponent    float64
	from        float64
	to          float64
	ramp        float64
	increment   float64
	lastOut     float64
}

func New(setting *Setting, releaseMode ReleaseMode, sr float64) *ADSR {
	return &ADSR{
		setting:     setting,
		releaseMode: releaseMode,
		stage:       idle,
		sr:          sr,
	}
}

func (adsr *ADSR) exponentFromShape(shape float64, direction float64) float64 {
	if shape == 0.0 {
		return 1.0
	}

	if (shape > 0.0 && direction > 0) || (shape < 0.0 && direction < 0) {
		return 1.0 / (1.0 + math.Abs(shape)*shapeMult)
	}

	return 1.0 + math.Abs(shape)*shapeMult
}

func (adsr *ADSR) nextStep(step step, useLastOut bool) {
	adsr.ramp = 0
	if useLastOut {
		adsr.from = adsr.lastOut
	} else {
		adsr.from = adsr.to
	}
	adsr.to = step.level
	adsr.stepCnt = int64(step.duration * adsr.sr * 0.001)
	adsr.increment = 1.0 / float64(adsr.stepCnt)
	adsr.exponent = adsr.exponentFromShape(step.shape, adsr.to-adsr.from)
}

func (adsr *ADSR) TriggerFull(duration float64, level float64, setting *Setting, releaseMode ReleaseMode) {
	adsr.setting = setting
	adsr.releaseMode = releaseMode
	adsr.duration = duration
	adsr.Trigger(level)
}

func (adsr *ADSR) TriggerWithDuration(duration float64, level float64) {
	adsr.duration = duration
	adsr.Trigger(level)
}

func (adsr *ADSR) Trigger(level float64) {
	nextStep, _ := adsr.setting.stepForStage(attack)
	adsr.stage = attack
	adsr.level = level
	adsr.releaseCnt = int64(adsr.duration * adsr.sr * 0.001)
	adsr.nextStep(nextStep, true)
}

func (adsr *ADSR) Release() {
	nextStep, _ := adsr.setting.stepForStage(release)
	adsr.stage = release
	adsr.nextStep(nextStep, true)
}

func (adsr *ADSR) Reset() {
	adsr.Trigger(adsr.level)
}

func (adsr *ADSR) Clear() {
	adsr.from = 0.0
	adsr.to = 0.0
	adsr.lastOut = 0.0
}

func (adsr *ADSR) Continuous() bool {
	return false
}

func (adsr *ADSR) Done() bool {
	return adsr.stage == idle
}

func (adsr *ADSR) ReleaseMode() ReleaseMode {
	return adsr.releaseMode
}

func (adsr *ADSR) Generate() float64 {
	if adsr.stage == idle {
		adsr.lastOut = 0.0
		goto output
	}

	if adsr.releaseMode == Duration && adsr.releaseCnt > 0 {
		adsr.releaseCnt--
		if adsr.releaseCnt == 0 {
			defer adsr.Release()
		}
	}

	if (adsr.releaseMode == NoteOff || adsr.releaseMode == Duration) && adsr.stage == sustain {
		adsr.lastOut = adsr.from
		goto output
	}

	adsr.lastOut = adsr.from + math.Pow(adsr.ramp, adsr.exponent)*(adsr.to-adsr.from)
	adsr.ramp += adsr.increment
	adsr.stepCnt--

	if adsr.stepCnt <= 0 {
		adsr.stage++
		if adsr.stage < idle {
			nextStep, nextStage := adsr.setting.stepForStage(adsr.stage)
			adsr.stage = nextStage
			adsr.nextStep(nextStep, false)
		}
	}

output:
	return adsr.lastOut * adsr.level
}
