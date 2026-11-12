# genny

Renders a piece described in JSON to a MIDI file, using the generators of this library.

    go run ./cmd/genny render examples/spec/groove.json -o groove.mid [-seed N]
    go run ./cmd/genny validate examples/spec/groove.json
    go run ./cmd/genny list [kind]

`render` and `validate` accept `-` as the file to read standard input. The seed comes from `-seed`, else from
the spec, else it is 1: the same spec and seed always give the same file. Exit status: 0 ok, 1 invalid spec
or I/O error, 2 usage error.

## Format

    {
      "tempo": 120, "steps_per_beat": 4, "seed": 7,
      "time_signature": {"num": 4, "den": 4},
      "tracks": [
        {"name": "lead", "channel": 0, "program": 80, "pitch_bend": false, "source": NODE}
      ]
    }

A node is `{"kind": NAME, "params": {...}}`. Generator kinds (melody, euclid, sieve, row, numseq, pulse,
tendency, choose, cloud, progression, chordmelody, notes) make a line of notes. Combinator kinds (concat,
layer/merge, repeat, transpose, invert, retrograde, stretch, shift, slice, quantize, canon, swing, humanize,
gain, thin) take child nodes in `item` or `items`. Time is in steps; pitches are numbers or names (`C4` is
60); scales are names or arrays of semitone steps. `genny list` prints every kind with its parameters and
defaults.

    {"kind": "concat", "params": {"items": [
      {"kind": "melody", "params": {"scale": "minor", "root": "D3", "length": 32}},
      {"kind": "transpose", "params": {"semitones": 7, "item":
        {"kind": "euclid", "params": {"pulses": 5, "steps": 8, "pitch": 36, "length": 32}}}}
    ]}}

A sieve is written as an expression: `mod@start` residue classes joined by `&`, `|`, `-`, `!` and
parentheses, for example `(3@0|5@1)&!2@0`.

## Errors

Unknown fields and wrong types are rejected, every parameter is range checked, and library panics are
caught. Errors quote the JSON path, and all broken tracks are reported at once:

    genny: bad.json: tracks[1].source.params.length: must be > 0, got 0
    tracks[2].source.params.items[1].params.pulses: must be in 0..8, got 9

## Adding a kind

All kinds are in one table, filled by `spec.Register` (see `spec/kinds_gen.go` and `spec/kinds_comb.go`). A
kind is a name, a category, a doc, the parameter docs shown by `genny list`, and a build function:
decode the params into a struct that holds the defaults, check them with the `check` helpers, build a
`spec.Line`. A kind for a new package (fugue, harmony, polyrhythm, microtonal) is a new file with an
`init()` that calls `Register`, nothing else changes.
