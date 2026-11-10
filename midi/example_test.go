package midi_test

import (
	"bytes"
	"fmt"

	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/sequence"
)

// Write a bar of notes as a MIDI file. WriteFile does the same to a file you can open in any DAW
func ExampleWrite() {
	voice := note.Voice{Pitch: note.Always[float64](sequence.NewLoop(60.0, 64.0, 67.0, 72.0))}
	var buf bytes.Buffer
	err := midi.Write(&buf, midi.File{
		Tempo:  100,
		Tracks: []midi.Track{{Name: "arp", Program: 4, Events: voice.Render(4)}},
	})
	fmt.Println(err, buf.Len(), "bytes, starts with", string(buf.Bytes()[:4]))
	// Output: <nil> 87 bytes, starts with MThd
}
