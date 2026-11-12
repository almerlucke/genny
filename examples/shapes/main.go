// shapes is a piece of 32 bars in D Dorian that is sculpted by shapers. Slow oscillators are the raw material
// and every shaper changes them into something musical:
//
//   - lead: two sines through a Legendre polynomial and a sine wavefolder whose gain grows over the piece;
//     funnel snaps the result to the scale; a Schmitt trigger on a third oscillator makes the rhythm;
//     the velocity comes from a Hann window per phrase, smoothstep and a tube saturator
//
//   - arp: a bit crushed saw (8 levels) through the funnel, the pitch glides in steps through a backlash (hysteresis)
//
//   - bass: sample and hold takes a value from an oscillator once a bar and the funnel snaps it to the chord roots
//
//   - hats: a full-wave rectified sine through a dead zone: dense and sparse in a cycle, with a bend curve for the velocity
//
//   - controllers: mod wheel from a phase distorted saw, expression from a one-pole smoothed gate,
//     the harmonics shaper makes the breathing of the dynamics
//
//     go run ./examples/shapes -seed 3 -mid shapes.mid
package main

import (
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/float/shape/shapers/bitcrush"
	"github.com/almerlucke/genny/float/shape/shapers/curve"
	"github.com/almerlucke/genny/float/shape/shapers/deadzone"
	"github.com/almerlucke/genny/float/shape/shapers/funnel"
	"github.com/almerlucke/genny/float/shape/shapers/harmonics"
	"github.com/almerlucke/genny/float/shape/shapers/hysteresis"
	"github.com/almerlucke/genny/float/shape/shapers/onepole"
	"github.com/almerlucke/genny/float/shape/shapers/phasedist"
	"github.com/almerlucke/genny/float/shape/shapers/polynomial"
	"github.com/almerlucke/genny/float/shape/shapers/rectify"
	"github.com/almerlucke/genny/float/shape/shapers/samplehold"
	"github.com/almerlucke/genny/float/shape/shapers/saturate"
	"github.com/almerlucke/genny/float/shape/shapers/smoothstep"
	"github.com/almerlucke/genny/float/shape/shapers/wavefold"
	"github.com/almerlucke/genny/float/shape/shapers/window"
	"github.com/almerlucke/genny/function"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

const (
	bars        = 32
	stepsPerBar = 16
	total       = bars * stepsPerBar
)

var names = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func name(p float64) string {
	n := int(math.Round(p))
	return fmt.Sprintf("%s%d", names[((n%12)+12)%12], n/12-1)
}

// osc is a sine of a period in steps, from -1 to 1
func osc(t, period, phase float64) float64 { return math.Sin(2 * math.Pi * (t/period + phase)) }

// saw is a ramp from 0 to 1 with a period in steps
func saw(t, period float64) float64 { return math.Mod(t/period, 1) }

func clamp01(x float64) float64 { return math.Max(0, math.Min(1, x)) }

// fromTo makes an event list from a per step function that returns false when there is no note
func main() {
	seed := flag.Uint64("seed", 1, "seed of the random stream (the funnel breaks ties with it)")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	dorian := scale.New(38, scale.Dorian) // D Dorian from D2
	lead := funnel.FromScale(dorian).WithTie(funnel.Random)
	arpSet := funnel.New(scaleNotes(dorian, 62, 86, []int{0, 2, 4, 6})) // a pentatonic subset: degrees 0 2 4 6 of the mode
	bassSet := funnel.New([]float64{26, 31, 33, 29})                    // D G A F

	// ---- shapers used below
	poly := polynomial.NewLegendre(5)
	schmitt := hysteresis.NewSchmitt(-0.25, 0.25, 0, 1)
	crush := bitcrush.New(3, 1, 1)
	backlash := hysteresis.NewPlay(0.35)
	hann := window.Hann()
	tube := saturate.NewTube(2.2, 0.3)
	rectifier := rectify.New(rectify.FullWave)
	dead := deadzone.NewDeadZone(0.35, deadzone.Jump)
	bend := curve.NewUnipolar(0.6)
	pd := phasedist.NewSaw(0.7)
	breath := harmonics.New([]float64{1, 0.5, 0.25}, []float64{0, 1, 2}, true)
	smooth := onepole.NewSamples(6)
	step := smoothstep.New(smoothstep.Smootherstep, 0, 1)
	var trigger bool
	hold := samplehold.New(function.New(func() bool { return trigger }))

	var leadEvents, arpEvents, bassEvents, hatEvents []note.Event
	var leadCC, arpCC []midi.Controller

	for t := 0; t < total; t++ {
		ft := float64(t)
		bar := t / stepsPerBar
		progress := ft / total

		// lead: two oscillators -> polynomial -> wavefolder (more folding as the piece goes on) -> scale
		raw := 0.55*osc(ft, 64, 0) + 0.25*osc(ft, 40, 0.1) + 0.2*osc(ft, 13, 0.3) // the fast third oscillator keeps the fold from flattening on an extreme
		shaped := wavefold.NewSine(1 + 2.5*progress).Shape(poly.Shape(raw))
		pitch := lead.Shape(62 + 14*shaped)

		// the rhythm: a Schmitt trigger on a third oscillator, two voices that interlock
		gate := schmitt.Shape(osc(ft, 19, 0.2)+0.2*osc(ft, 7, 0)) > 0.5

		// velocity: a Hann window over a phrase of four bars, S-curved, with the tube's saturation
		phrasePos := math.Mod(ft/(4*stepsPerBar), 1)
		vel := 0.25 + 0.65*tube.Shape(step.Shape(hann.Shape(phrasePos)))
		vel = math.Max(0.2, math.Min(1, vel))
		if gate && bar >= 2 {
			leadEvents = append(leadEvents, note.Event{Pitch: pitch, Velocity: vel, Start: ft, Duration: 1.5 + 2*bend.Shape(clamp01(0.5+0.5*osc(ft, 33, 0)))})
		}

		// arp: a crushed saw, pitch glides in steps through the backlash, then snapped to the set
		if t%2 == 0 && bar >= 6 && bar < 28 {
			level := crush.Shape(2*saw(ft, 48) - 1) // 8 levels
			glide := backlash.Shape(0.6*level + 0.4*osc(ft, 80, 0))
			arpEvents = append(arpEvents, note.Event{Pitch: arpSet.Shape(62 + 22*(glide+1)/2), Velocity: 0.35 + 0.25*clamp01(osc(ft, 24, 0)+0.5), Start: ft, Duration: 1.6})
			if t%8 == 0 { // pan from a sine
				arpCC = append(arpCC, midi.Controller{Step: ft, Number: 10, Value: int(64 + 50*osc(ft, 96, 0))})
			}
		}

		// bass: sample and hold once a bar, an oscillator snapped to the roots
		trigger = t%stepsPerBar == 0
		heldRaw := hold.Shape(26 + 8*osc(ft/stepsPerBar*3, 7, 0))
		if bar >= 4 && trigger {
			bassEvents = append(bassEvents, note.Event{Pitch: bassSet.Shape(heldRaw), Velocity: 0.7, Start: ft, Duration: float64(stepsPerBar) - 1})
		}

		// hats: full-wave rectified sine through a dead zone: dense and sparse over a cycle of two bars
		r := rectifier.Shape(osc(ft, 32, 0))
		if h := dead.Shape(r); h > 0 && bar >= 8 {
			hatEvents = append(hatEvents, note.Event{Pitch: 42, Velocity: 0.3 + 0.6*bend.Shape(clamp01(h)), Start: ft, Duration: 0.5})
		}

		// controllers: mod wheel from a phase distorted saw (one cycle a bar), expression from the smoothed gate
		// with a breathing dynamic from the harmonics shaper
		if t%2 == 0 {
			leadCC = append(leadCC, midi.Controller{Step: ft, Number: 1, Value: int(127 * clamp01((pd.Shape(saw(ft, stepsPerBar))+1)/2))})
		}
		g := 0.0
		if gate {
			g = 1
		}
		expr := 0.55 + 0.3*smooth.Shape(g) + 0.15*breath.Shape(osc(ft, 128, 0))
		if t%4 == 0 {
			leadCC = append(leadCC, midi.Controller{Step: ft, Number: 11, Value: int(127 * clamp01(expr))})
		}
	}

	tracks := []midi.Track{
		{Name: "lead", Channel: 0, Program: 11, Events: leadEvents, Controllers: leadCC},
		{Name: "arp", Channel: 1, Program: 46, Events: arpEvents, Controllers: arpCC},
		{Name: "bass", Channel: 2, Program: 38, Events: bassEvents},
		{Name: "hats", Channel: 9, Events: hatEvents},
	}

	for bar := 0; bar < bars; bar += 4 {
		fmt.Printf("bar %2d lead:", bar+1)
		n := 0
		for _, e := range leadEvents {
			if int(e.Start)/stepsPerBar >= bar && int(e.Start)/stepsPerBar < bar+4 && n < 10 {
				fmt.Printf(" %s", name(e.Pitch))
				n++
			}
		}
		fmt.Println()
	}
	for _, t := range tracks {
		fmt.Printf("%-6s %4d notes, %3d controller values\n", t.Name, len(t.Events), len(t.Controllers))
	}

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: 100, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}

// scaleNotes lists the notes of a scale between lo and hi that sit on the given degrees of the scale
func scaleNotes(s *scale.Scale, lo, hi float64, degrees []int) []float64 {
	var out []float64
	for i := -20; i < 40; i++ {
		ok := false
		for _, d := range degrees {
			if ((i%s.Len())+s.Len())%s.Len() == d {
				ok = true
			}
		}
		if p := s.Degree(i); ok && p >= lo && p <= hi {
			out = append(out, p)
		}
	}
	return out
}
