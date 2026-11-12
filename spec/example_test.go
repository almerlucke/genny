package spec_test

import (
	"encoding/json"
	"fmt"

	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/spec"
)

func Example() {
	s, err := spec.ParseBytes([]byte(`{
	  "tempo": 120, "seed": 3,
	  "tracks": [
	    {"name": "kick", "channel": 9, "source": {"kind": "euclid", "params": {"steps": 8, "pulses": 3, "pitch": 36, "length": 16}}},
	    {"name": "bass", "channel": 1, "source": {"kind": "transpose", "params": {"semitones": -12, "item": {
	      "kind": "melody", "params": {"scale": "minor", "root": "D3", "length": 16}}}}}
	  ]}`))
	if err != nil {
		panic(err)
	}
	res, err := s.Render(nil)
	fmt.Println(len(res.File.Tracks), "tracks", err)

	// every error quotes the JSON path of the problem
	s, _ = spec.ParseBytes([]byte(`{"tracks": [{"source": {"kind": "melody", "params": {"length": 0}}}]}`))
	_, err = s.Render(nil)
	fmt.Println(err)
	// Output:
	// 2 tracks <nil>
	// tracks[0].source.params.length: must be > 0, got 0
}

func ExampleRegister() {
	// a new kind is one table entry: here a single note of a given pitch
	spec.Register(spec.Kind{
		Name: "drone", Category: "generator", Doc: "one long note",
		Params: []spec.Param{{Name: "pitch", Type: "pitch", Default: "36", Doc: "the note"}},
		Build: func(b *spec.Builder, path string, params json.RawMessage) (spec.Line, error) {
			p := struct {
				Pitch float64 `json:"pitch"`
			}{Pitch: 36}
			if err := b.Decode(path, params, &p); err != nil {
				return spec.Line{}, err
			}
			return spec.Line{Length: 16, Events: []note.Event{{Pitch: p.Pitch, Velocity: 0.8, Duration: 16}}}, nil
		},
	})
	k, _ := spec.Lookup("drone")
	fmt.Println(k.Name, "-", k.Doc)
	// Output: drone - one long note
}
