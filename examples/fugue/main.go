// fugue is a fugue in four voices in D minor, about 75 seconds, built from a subject of two bars with package fugue:
//
//   - an exposition: the voices enter from the top down, subject and answer in turn (a tonal answer, so the
//     answer starts a fourth above the first note, not a fifth), with a countersubject written by package
//     counterpoint (second species, invertible at the octave) and free lines in the other voices
//   - an episode: a sequence of the head of the subject, with the head of the countersubject, down by step
//   - a middle entry in F major (the relative major), with the countersubject now below the subject
//   - an episode in the lower voices and a middle entry in the bass in G
//   - an episode on the way home, and a stretto: all four voices enter a bar apart
//   - a close: a dominant and then a tonic pedal, with the leading tone raised
//
// Every voice is on its own MIDI channel and the entries are checked with Fugue.Check.
//
//	go run ./examples/fugue -seed 3 -mid fugue.mid
package main

import (
	"flag"
	"fmt"
	"log"
	"math"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/fugue"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/scale"
)

var voiceNames = []string{"soprano", "alto", "tenor", "bass"}
var programs = []int{68, 71, 41, 70} // oboe, clarinet, viola, bassoon

func main() {
	seed := flag.Int("seed", 3, "seed of the random stream: the same seed gives the same piece")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(uint64(*seed))

	// D minor (natural), the subject in degrees: D A Bb G | E A G F E D, two bars of sixteenth steps
	key := scale.New(62, scale.Minor)
	subject := fugue.NewSubject(
		[]int{0, 4, 5, 3, 1, 4, 3, 2, 1, 0},
		[]float64{8, 4, 2, 2, 4, 2, 2, 2, 2, 4},
	)
	cfg := fugue.Config{Answer: fugue.Tonal}

	f, err := fugue.Exposition(subject, 4, key, cfg)
	if err != nil {
		fmt.Println("countersubject:", err)
	}
	answer := fugue.Answer(subject, key, fugue.Tonal)
	fmt.Println("subject (degrees):", subject.Degrees)
	fmt.Println("answer  (degrees):", answer.Degrees)
	fmt.Println("countersubject   :", f.Counter.Degrees, "invertible:", f.CounterInvertible)

	// an episode of the head, down by step, that ends where the sequence would be on F (degree 2)
	arrival := f.Episode(fugue.EpisodeConfig{Voices: []int{0, 1}, Repeats: 5, Step: -1})
	fmt.Println("episode 1 leads to degree", arrival%7)
	f.Middle(1, fugue.KindSubject, 2, true) // F major, the countersubject above

	f.Episode(fugue.EpisodeConfig{Voices: []int{2, 3}, Repeats: 3, Step: -1, Shift: 6})
	f.Middle(3, fugue.KindSubject, 3, true) // G dorian in the bass

	f.Episode(fugue.EpisodeConfig{Voices: []int{0, 1}, Repeats: 3, Step: -1, Shift: 3})
	dists := fugue.StrettoDistances(subject, key, fugue.Tonal, 4, 8)
	fmt.Println("stretto distances that work (steps):", dists)
	if err := f.Stretto(16, nil); err != nil {
		log.Fatal("stretto: ", err)
	}
	f.Close(fugue.CloseConfig{Bars: 4, LeadingTone: true})

	fmt.Println("\nsections:")
	for _, s := range f.Sections {
		fmt.Printf("  %-13s bars %2d to %2d\n", s.Name, f.Bar(s.Start), f.Bar(s.End-1e-6))
	}
	fmt.Println("\nentries:")
	for _, e := range f.Entries {
		fmt.Printf("  bar %2d  %-7s %-7s %-12s shift %+d degrees\n", f.Bar(e.Start), voiceNames[e.Voice], e.Kind, e.Section, e.Shift)
	}
	problems := f.Check()
	for _, p := range problems {
		fmt.Println("PROBLEM:", p)
	}
	fmt.Printf("\nentries checked: %d, problems: %d\n", len(f.Entries), len(problems))

	// 92 beats per minute, a bar is 4 beats of 4 steps
	const tempo = 92.0
	seconds := f.End() / 4 * 60 / tempo
	fmt.Printf("%d bars, %.0f seconds\n", int(math.Ceil(f.End()/16)), seconds)

	if *midiFile != "" {
		var tracks []midi.Track
		for v, events := range f.Voices {
			tracks = append(tracks, midi.Track{Name: voiceNames[v], Channel: v, Program: programs[v], Events: events})
		}
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: tempo, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("MIDI written to", *midiFile)
	}
}
