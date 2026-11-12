// learnstyle reads a melody from a MIDI file, learns its style with an n-gram (package ngram/style) and
// plays a new melody, four times as long, in that style. The new melody is written next to the original
// so you can compare them.
//
// Without -in it builds a hymn tune (Ode to Joy, by Beethoven, in the public domain) as note numbers,
// writes it to a temporary MIDI file and reads it back with midi.ReadFile, so the import is part of the
// run. With -in it learns from the track with most notes of any MIDI file (for instance one made by another
// example). The style is learned as pitch intervals and duration ratios, so what is made follows the
// shape of the tune, not its key.
//
// With -quantize the notes of an imported file are moved to the nearest notes of a scale (scale.Quantize), to show
// how an existing file can be transformed; the new melody is quantized to the same scale. Without -in the
// new melody is quantized to C major, the key of the tune.
//
//	go run ./examples/learnstyle -seed 3 -mid learn.mid
//	go run ./examples/learnstyle -in piece.mid -quantize minor -root 57 -mid piece_learned.mid
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/ngram/style"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

var scales = map[string][]float64{
	"major": scale.Major, "minor": scale.Minor, "harmonicminor": scale.HarmonicMinor, "dorian": scale.Dorian,
	"mixolydian": scale.Mixolydian, "pentatonic": scale.MajorPentatonic, "minorpentatonic": scale.MinorPentatonic,
	"blues": scale.Blues, "wholetone": scale.WholeTone,
}

var names = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func name(p float64) string {
	n := int(math.Round(p))
	s := fmt.Sprintf("%s%d", names[((n%12)+12)%12], n/12-1)
	if cents := int(math.Round((p - float64(n)) * 100)); cents != 0 {
		s += fmt.Sprintf("%+dc", cents)
	}
	return s
}

// odeToJoy is the tune as note numbers (60 is middle C) and durations in sixteenths
func odeToJoy() []note.Event {
	type n struct{ pitch, dur float64 }
	var line1 = []n{{64, 4}, {64, 4}, {65, 4}, {67, 4}, {67, 4}, {65, 4}, {64, 4}, {62, 4}, {60, 4}, {60, 4}, {62, 4}, {64, 4}}
	line2 := append(append([]n{}, line1...), n{62, 6}, n{62, 2}, n{62, 8}) // ... D. D D (half)
	line1 = append(line1, n{64, 6}, n{62, 2}, n{62, 8})
	bridge := []n{{62, 4}, {62, 4}, {64, 4}, {60, 4}, {62, 4}, {64, 2}, {65, 2}, {64, 4}, {60, 4},
		{62, 4}, {64, 2}, {65, 2}, {64, 4}, {62, 4}, {60, 4}, {62, 4}, {55, 8}}
	line3 := append(append([]n{}, line1[:12]...), n{64, 6}, n{62, 2}, n{60, 8})

	var tune []n
	tune = append(tune, line1...)
	tune = append(tune, line2...)
	tune = append(tune, bridge...)
	tune = append(tune, line3...)

	var events []note.Event
	t := 0.0
	for _, x := range tune {
		events = append(events, note.Event{Pitch: x.pitch, Velocity: 0.8, Start: t, Duration: x.dur})
		t += x.dur
	}
	return events
}

// end is the time at which the last note stops
func end(events []note.Event) float64 {
	e := 0.0
	for _, ev := range events {
		e = max(e, ev.Start+ev.Duration)
	}
	return e
}

func printNotes(title string, phrase []style.Note) {
	var sb strings.Builder
	for i, n := range phrase {
		if i == 40 {
			fmt.Fprintf(&sb, "... (%d more)", len(phrase)-i)
			break
		}
		fmt.Fprintf(&sb, "%s:%g ", name(n.Pitch), math.Round(n.Duration*100)/100)
	}
	fmt.Printf("%s (%d notes)\n  %s\n", title, len(phrase), sb.String())
}

func main() {
	seed := flag.Uint64("seed", 3, "seed of the random stream: the same seed gives the same melody")
	in := flag.String("in", "", "learn from this MIDI file (default: the tune Ode to Joy, made in code)")
	midiFile := flag.String("mid", "", "write the original and the new melody to this MIDI file")
	order := flag.Int("order", 2, "how many notes the chain looks back")
	quantize := flag.String("quantize", "", "move the notes to this scale: "+strings.Join(sortedKeys(), ", "))
	root := flag.Int("root", 60, "root note of the quantize scale (60 is middle C)")
	flag.Parse()
	genny.Seed(*seed)

	var key *scale.Scale
	if *quantize != "" {
		steps, ok := scales[*quantize]
		if !ok {
			log.Fatalf("unknown scale %q, choose one of %s", *quantize, strings.Join(sortedKeys(), ", "))
		}
		key = scale.New(float64(*root), steps)
	}

	// ---- the example: made in code and written to a file, or given. Either way it is read with midi.ReadFile
	path := *in
	if path == "" {
		dir, err := os.MkdirTemp("", "learnstyle")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(dir)
		path = filepath.Join(dir, "ode.mid")
		err = midi.WriteFile(path, midi.File{
			Tempo:          100,
			Tracks:         []midi.Track{{Name: "Ode to Joy", Program: 73, Events: odeToJoy()}},
			TimeSignatures: []midi.TimeSignature{{Num: 4, Den: 4}},
		})
		if err != nil {
			log.Fatal(err)
		}
		if key == nil {
			key = scale.New(60, scale.Major)
		}
	}
	file, err := midi.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("read %s: tempo %g, %d tracks, %g steps per beat\n", filepath.Base(path), file.Tempo, len(file.Tracks), file.StepsPerBeat)

	// the track with most notes (not the drums) is the example
	src := -1
	for i, t := range file.Tracks {
		if t.Channel != 9 && (src < 0 || len(t.Events) > len(file.Tracks[src].Events)) {
			src = i
		}
	}
	if src < 0 || len(file.Tracks[src].Events) <= *order {
		log.Fatalf("no track with more than %d notes to learn from", *order)
	}

	// ---- transform the imported file: every note to the nearest note of the scale
	if *in != "" && key != nil {
		for i := range file.Tracks {
			for j := range file.Tracks[i].Events {
				file.Tracks[i].Events[j].Pitch = key.Quantize(file.Tracks[i].Events[j].Pitch)
			}
		}
	}
	example := file.Tracks[src]
	fmt.Printf("learning from track %q: %d notes, %g steps long\n", example.Name, len(example.Events), end(example.Events))
	printNotes("original", style.Melody(example.Events))

	// ---- learn and play
	st := style.Learn(example.Events, *order, style.Intervals(), style.Ratios())
	phrase := st.Fill(4 * end(example.Events))
	if key != nil {
		for i := range phrase {
			phrase[i].Pitch = key.Quantize(phrase[i].Pitch)
		}
	}
	printNotes("new melody in the style", phrase)

	if *midiFile != "" {
		used := map[int]bool{9: true}
		for _, t := range file.Tracks {
			used[t.Channel] = true
		}
		ch := 0
		for used[ch] {
			ch = (ch + 1) % 16
		}
		generated := midi.Track{
			Name: "learned", Channel: ch, Program: example.Program, PitchBend: example.PitchBend,
			Events: style.Events(phrase, 0, st.Velocity),
		}

		out := file
		if *in == "" || key == nil {
			out.Tracks = []midi.Track{example}
		}
		out.Tracks = append(out.Tracks, generated)
		if err := midi.WriteFile(*midiFile, out); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}

func sortedKeys() []string {
	var keys []string
	for k := range scales {
		keys = append(keys, k)
	}
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
