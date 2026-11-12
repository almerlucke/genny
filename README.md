# genny

Generators for making music with Go. A generator hands out values one at a time: a melody, a rhythm,
a modulation signal. You combine small generators into bigger ones (a mutating phrase on a Euclidean
rhythm, quantized to a scale, swung, written to a MIDI file) and you can seed the whole thing, so the
same seed gives the same piece.

```go
genny.Seed(7)

key := scale.New(57, scale.Minor)                                  // A minor
phrase := mutate.New([]int{7, 9, 11, 9, 8, 7, 4, 7}, mutate.Config[int]{SwapProb: 0.6})
lead := note.Voice{
	Pitch:    rest.NewGated[float64](scale.NewDegrees(key, phrase), euclid.New(16, 7, 0)),
	Velocity: groove.NewAccent(4, 4, 4, 4),
	Offset:   groove.NewSwing(0.15),
}
midi.WriteFile("lead.mid", midi.File{
	Tempo:  110,
	Tracks: []midi.Track{{Name: "lead", Events: lead.Render(64)}},
})
```

`go run ./examples/music -seed 7 -mid piece.mid` is a bigger version of this: chords, bass and lead.

## The generator

Everything implements one small interface:

```go
type Generator[T any] interface {
	Generate() T      // the next value
	Continuous() bool // false if Done and Reset are meaningful: the generator can end
	Done() bool       // true when a non-continuous generator has ended
	Reset()           // start again
}
```

Wrappers and combinators take generators and return generators, so they nest freely.

**Random numbers** all come from one stream. `genny.Seed(n)` makes every generator repeatable: after the
same seed the same calls give the same values. Without a seed every run is different. Generators that
draw at creation (`noise`, `lfo`, `bucket`, `automaton.NewRandom`) must be created after the seed.

**Conventions.** A constructor panics on arguments that can never work (`euclid.New(0, ...)`, a scale
with descending steps) so the mistake shows at setup. `Generate` never panics. Empty input is allowed:
a `sequence`, `bucket`, `and`, `or` or `arpeggio` without values is done at once and generates the zero
value. Where a value has a natural limit it is clamped instead (more pulses than steps, a probability
above 1).

## Packages

### Building blocks

| Package | What it does |
|---|---|
| `sequence` | Plays a list of values once or in a loop |
| `bucket` | Hands out values at random, or one by one in shuffled order |
| `arpeggio` | Up, converge, alternate and random patterns, with mirroring |
| `constant`, `function` | A fixed value, a Go function as a generator |
| `markov` | A Markov chain over states, `ProbabilityState` for weighted transitions |
| `walk` | A random walk through an N dimensional container |
| `and`, `or` | Play generators one after the other, or pick one |
| `combine`, `flatten`, `unwrap`, `template` | Combine generators into slices, maps, or flatten them |
| `transform`, `conv` | Change the values (`T` to `T`, or to another type) |
| `repeat`, `continuous` | Give a continuous generator an end, or a finished one a loop |

### Rhythm and time

| Package | What it does |
|---|---|
| `euclid` | Euclidean rhythms (Bjorklund): 5 pulses in 8 steps is x.xx.xx. |
| `rest` | Rests: `Gated` by a gate generator, or `Probability` thins a melody out |
| `pulse` | Metric hierarchy (Barlow style): thin a rhythm by importance, fade it in and out with a level generator |
| `clock` | `Divider` (every nth step) and `Poly` (cycles of different lengths) |
| `groove` | `Swing`, `Humanize`, `Jitter` for timing, `Accent` for velocities by meter (7/8 as 2+2+3) |
| `ratchet` | Repeats a step several times: a roll. Random or from a pattern |
| `section` | Plays parts for N steps each: A for 4 bars, B for 8, and around |
| `automaton` | Cellular automata as patterns and gates: one dimensional (rule 30, 90, 110) and Game of Life (B/S rules) |

### Pitch and harmony

| Package | What it does |
|---|---|
| `scale` | Scales (modes, Messiaen's modes, pentatonic, blues, just intonation, n equal divisions, Scala `.scl` files), quantizing, `Walk` over degrees |
| `chord` | Chords, inversions, diatonic chords from a scale, a `VoiceLeader` for smooth voice leading, spectral chords (harmonic, FM, ring modulation) |
| `line` | Transformations of whole lines of note events: canon, inversion, retrograde, augmentation, loop, thin |
| `counterpoint` | First species counterpoint against a cantus firmus, with a rule checker (second species experimental) |
| `melody` | A melody over chords: chord tones on strong beats, scale tones between |
| `progression` | Chord progressions as a Markov chain (pop, classical, jazz, blues), or your own table |
| `row` | Tone rows: retrograde, inversion, transposition, the twelve tone matrix, permutations and all-interval rows |
| `ngram` | Learns the style of an example melody and plays new music in it |

### Variation and structure

| Package | What it does |
|---|---|
| `mutate` | A phrase that changes a little every time it is played |
| `choose` | Picks between generators with weights, picks list elements by a distribution, weight sets that change over time |
| `lsystem` | Lindenmayer systems: self-similar phrases from rewriting rules |
| `numseq` | Fibonacci, Collatz, Thue-Morse and the primes |
| `grammar` | Context-free rewrite grammars with weighted alternatives: phrases and bars that are derived again every cycle |
| `stochastic` | Xenakis style: Poisson onsets, clouds with a changing density, free stochastic music |
| `evolve` | A genetic algorithm for phrases: fitness for smoothness, contour, scale; a player that improves a phrase as it plays |
| `isorhythm`, `phase` | Ars nova talea against color of different lengths; Reich style phasing of a pattern against itself |
| `constrain` | Filters that redraw until a rule is met: allowed or blocked intervals, no repetitions, no duplicates in a group; constrained melodic walks |
| `sieve` | Xenakis sieves: residue classes with union, intersection and complement, as pitch sets, gates and filters |
| `split` | Recursive rhythm splitting: a duration divided into branches until a lower limit |
| `listtools` | Jump, stutter, group, grow, gather-until, plus-min, positive-negative |

### Numbers, signals and modulation (`float/...`)

| Package | What it does |
|---|---|
| `float/phasor`, `float/line`, `float/ramp` | Phase ramps, lines and curves between values |
| `float/lfo` | LFO: sine, triangle, saw, square, sample-and-hold, smooth random, and a slew limiter |
| `float/noise` | Smooth (Perlin style) noise with octaves |
| `float/rand`, `float/dist` | Uniform random values; Gaussian, exponential, beta, Cauchy and Weibull ones, Lehmer and pink noise |
| `float/tendency` | Tendency masks: values or choices between two moving bounds, with beta variants, two-mask rejection, anchor and spray |
| `float/rwalk` | A random walk between limits (clip, bounce, jump) |
| `float/lissajous` | Two related modulation signals |
| `float/iter` + `updaters/chaos` | Iterated maps and attractors: Verhulst, Hénon, Aronson, Lorenz, Rössler |
| `float/envelopes/adsr` | ADSR envelopes |
| `float/interp` | Reads a frame generator at another speed, with interpolation |
| `float/shape` + `shapers/...` | Waveshaping: lookup tables, clipping, mirroring, quantizing, Chebyshev and others |
| `shapers/wavefold`, `saturate`, `rectify`, `curve`, `smoothstep`, `deadzone`, `harmonics`, `polynomial` | Wavefolders (sine, triangle), tanh/atan/logistic/tube saturation, rectifiers, bend curves, S-curves, dead zone and expander, additive harmonics, polynomials (Legendre) |
| `shapers/phasedist`, `window` | Casio CZ phase distortion (saw, square, pulse, resonant); window functions (Hann, Hamming, Blackman, Tukey, Gaussian, Kaiser, Welch) |
| `shapers/bitcrush`, `hysteresis`, `onepole`, `samplehold`, `funnel` | Shapers with memory (bit and sample rate reduction, Schmitt trigger and backlash, one-pole filter, sample and hold) and snapping to a scale, a sieve or a set of values |
| `float/export/plot` | Plots a generator to a PNG, and a piano roll of note events |

### Notes and output

| Package | What it does |
|---|---|
| `note` | A `Voice` turns generators (pitch with rests, velocity, length, timing offset, ratchets) into note events |
| `midi` | Reads and writes Standard MIDI Files: tempo and meter changes, controllers, program changes, pitch bend for microtonal pitches |
| `ngram/style` | Learns pitches and durations from imported events (as intervals and ratios, so it transposes) and plays new music in that style |

## Examples

Every package has runnable examples in its `example_test.go` (`go doc` shows them, `go test` checks them).
There are also programs that play a piece as text and write a MIDI file (`-mid`):

```sh
go run ./examples/music -seed 3 -mid out.mid   # a piece as text and as a MIDI file
go run ./examples/tendency -seed 3 -mid t.mid    # tendency masks and distributions
go run ./examples/rhythmtools -seed 3 -mid r.mid # sieves, pulse hierarchy, splitting
go run ./examples/constraints -seed 3 -mid c.mid # grammar, constrained walks, spectral chords
go run ./examples/suite -seed 3 -mid s.mid       # all of them in one piece of 32 bars
go run ./examples/rexregum -seed 7 -mid r.mid    # Rex Regum: a piece on biblical numbers (3, 7, 12, 153)
go run ./examples/shapes -seed 3 -mid sh.mid       # a piece sculpted by shapers (wavefold, funnel, bitcrush, ...)
go run ./examples/counterpoint -seed 3 -mid cp.mid # counterpoint, canon, isorhythm, phasing
go run ./examples/evolve -seed 3 -mid e.mid     # Game of Life, stochastic cloud, evolved melody, Bohlen-Pierce
go run ./examples/learnstyle -seed 3 -mid l.mid  # reads a MIDI file, learns its style, plays new music in it
(cd examples/plotmusic && go run . plots)              # PNG plots of the generators, to see that they behave
```

## Modules

The library itself has no dependencies. The plotting code (`float/export/plot`, `examples/plot`,
`examples/plotmusic`) uses gonum and lives in modules of its own; a `go.work` file ties them together
for local work.

## Tests

```sh
go test ./...                                      # the library
(cd float/export/plot && go test ./...)            # the plotting module
```
