// harmony writes a chorale in four voices: functional harmony in G major that moves to D major (the
// dominant) through a pivot chord and returns, realised in soprano, alto, tenor and bass with the
// rules of voice leading. Each voice has its own MIDI channel. It prints the Roman numerals per bar,
// the notes, and every rule the realisation had to break.
//
//	go run ./examples/harmony                       a text score
//	go run ./examples/harmony -seed 7 -mid out.mid  another chorale, also as a MIDI file
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/harmony"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
)

const stepsPerBar = 16 // one chord per bar, four quarter notes

var names = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func name(p int) string { return fmt.Sprintf("%s%d", names[p%12], p/12-1) }

func main() {
	seed := flag.Uint64("seed", 1, "seed of the random stream: the same seed gives the same chorale")
	midiFile := flag.String("mid", "", "write the chorale to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	g := harmony.Key{Tonic: 7}
	d := g.Dominant()
	gen := harmony.New(harmony.Config{
		Key: g, PhraseLength: 4,
		SeventhProb: 0.35, SecondaryProb: 0.3, BorrowedProb: 0.1,
	})

	// G major: half cadence, authentic cadence, then through a pivot to D major, a phrase there, and
	// home again with a direct modulation (the dominant seventh of G) to the last phrase
	var chords []harmony.Chord
	add := func(cs ...harmony.Chord) { chords = append(chords, cs...) }
	add(gen.Phrase(4, harmony.Half)...)
	add(gen.Phrase(4, harmony.Authentic)...)
	add(harmony.Modulate(g, d)...)
	gen.SetKey(d)
	add(gen.Phrase(4, harmony.Authentic)...)
	add(harmony.Modulate(d, g)...)
	gen.SetKey(g)
	add(gen.Phrase(4, harmony.Authentic)...)

	voicings, exceptions := harmony.Realize(chords, harmony.VoiceConfig{Variety: 2})

	// each voice is a track on its own channel: a note is held while the pitch stays the same
	voiceNames := []string{"soprano", "alto", "tenor", "bass"}
	tracks := make([]midi.Track, 4)
	for v := range tracks {
		var events []note.Event
		for i, vo := range voicings {
			start := float64(i * stepsPerBar)
			length := float64(stepsPerBar) * 0.97
			if i == len(voicings)-1 {
				length = 2 * stepsPerBar // hold the last chord
			}
			if n := len(events); n > 0 && int(events[n-1].Pitch) == vo.Notes[v] && events[n-1].Start+events[n-1].Duration >= start-1 {
				events[n-1].Duration = start + length - events[n-1].Start
				continue
			}
			events = append(events, note.Event{Pitch: float64(vo.Notes[v]), Velocity: 0.7, Start: start, Duration: length})
		}
		tracks[v] = midi.Track{Name: voiceNames[v], Channel: v, Program: 52, Events: events} // choir aahs
	}

	fmt.Printf("%d bars, seed %d\n", len(voicings), *seed)
	for i, v := range voicings {
		c := v.Chord
		extra := ""
		if i > 0 && voicings[i-1].Chord.Key != c.Key {
			extra = "  <- " + c.Key.String()
		}
		fmt.Printf("bar %2d  %-8s %-6s %s  %-4s %-4s %-4s %-4s%s\n", i+1, c.Numeral, c.Name(), c.Function,
			name(v.Notes[0]), name(v.Notes[1]), name(v.Notes[2]), name(v.Notes[3]), extra)
	}
	if len(exceptions) == 0 {
		fmt.Println("no rule had to be broken")
	}
	for _, e := range exceptions {
		fmt.Println("exception:", e)
	}

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: 60, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}
