package fugue

import (
	"errors"
	"math"

	"github.com/almerlucke/genny/counterpoint"
	"github.com/almerlucke/genny/note"
	"github.com/almerlucke/genny/scale"
)

// ErrStretto is returned by Stretto when the entries would form dissonances on the beats at that distance
// (see StrettoDistances for the distances that work) or the arguments make no stretto
var ErrStretto = errors.New("fugue: the stretto does not work at this distance")

// counter is the countersubject for an entry of this kind
func (f *Fugue) counter(k Kind) Subject {
	if k == KindAnswer {
		return f.CounterAnswer
	}
	return f.Counter
}

func (f *Fugue) fifth() int { return f.Key.DegreeOf(f.Key.Root() + 7) }

// Exposition starts a fugue: the voices enter one after the other, each with the subject or the answer in
// turn, the first with the subject. While a voice enters, the voice that entered before it plays the
// countersubject, and the voices before that play a free line in first species against the entry
// (when one is found that is consonant with the countersubject on the beats and does not cross it;
// otherwise the voice rests). With one voice there is only the subject. Entries start on a barline, one
// subject length (rounded up to whole bars) after each other, and the voices sit around Config.Centers.
//
// The error is nil, or ErrNotInvertible / ErrNoCountersubject / ErrWholeSemitones from Countersubject:
// the fugue is returned in any case and a missing countersubject makes the voices rest. It uses the
// random stream, so seed first. It panics if voices < 1, if the key is nil or the subject has no notes
func Exposition(subject Subject, voices int, key *scale.Scale, cfg Config) (*Fugue, error) {
	if voices < 1 || key == nil || len(subject.Degrees) == 0 || len(subject.Degrees) != len(subject.Durations) {
		panic("fugue: an exposition needs a subject, a key and at least one voice")
	}
	f := &Fugue{Key: key, Config: cfg, Subject: subject, Voices: make([][]note.Event, voices)}
	f.Answer = Answer(subject, key, cfg.Answer)
	var err error
	if voices > 1 {
		f.Counter, err = Countersubject(subject, key, cfg)
		f.CounterInvertible = err == nil
		if errors.Is(err, ErrNoCountersubject) || errors.Is(err, ErrWholeSemitones) {
			f.Counter = Subject{}
		} else {
			f.CounterAnswer, _ = CounterAnswer(subject, f.Counter, key, cfg.Answer)
		}
	}
	order, centers := cfg.order(voices), cfg.centers(voices)
	bar := cfg.bar()
	entryLen := math.Ceil(subject.Length()/bar-1e-9) * bar

	for k, v := range order {
		kind := Kind(k % 2)
		base := f.base(kind)
		shift := fit(key, base.Degrees, 0, centers[v])
		start := float64(k) * entryLen
		entry := base.Events(key, start, shift)
		f.add(v, append([]note.Event(nil), entry...))
		f.Entries = append(f.Entries, Entry{Voice: v, Start: start, Kind: kind, Shift: shift, Section: "exposition"})
		if k == 0 || len(f.Counter.Degrees) == 0 {
			continue
		}

		cv := order[k-1]
		csShift := shift
		if cv > v { // the countersubject is written above; below the entry it is moved down an octave
			csShift -= key.Len()
		}
		cs := f.counter(kind).Events(key, start, csShift)
		f.add(cv, append([]note.Event(nil), cs...))

		// the free voices, the nearest to the countersubject first, each consonant with the lines before it
		placed := []placedLine{{cv, cs}}
		for i := k - 2; i >= 0; i-- {
			fv := order[i]
			if line := f.freeLine(entry, v, fv, centers[fv], placed); line != nil {
				f.add(fv, line)
				placed = append(placed, placedLine{fv, line})
			}
		}
	}
	f.section("exposition", 0)
	return f, err
}

type placedLine struct {
	voice int
	notes []note.Event
}

// freeLine writes a first species line in voice fv against the entry in voice ev. At every note of the entry
// it is consonant with the lines already placed and on its own side of them (above the lines of the voices
// below it; a unison is allowed). It returns nil if there is no such line within the budget
func (f *Fugue) freeLine(entry []note.Event, ev, fv int, center float64, placed []placedLine) []note.Event {
	cantus := make([]int, len(entry))
	for i, e := range entry {
		if math.Abs(e.Pitch-math.Round(e.Pitch)) > 1e-9 {
			return nil
		}
		cantus[i] = int(math.Round(e.Pitch))
	}
	side := counterpoint.Above
	if fv > ev {
		side = counterpoint.Below
	}
	filter := func(j, p int) bool {
		for _, o := range placed {
			q, ok := sounding(o.notes, entry[j].Start)
			if !ok {
				continue
			}
			if fp := float64(p); !counterpoint.Consonant(p-int(math.Round(q))) || (fp > q && fv > o.voice) || (fp < q && fv < o.voice) {
				return false
			}
		}
		return true
	}
	cfg := counterpoint.Config{Side: side, Scale: f.Key, Low: int(center) - 12, High: int(center) + 12, MaxDistance: 40, Budget: 40000, Filter: filter}
	for range 5 {
		cp, err := counterpoint.Generate(cantus, cfg)
		if err != nil {
			continue
		}
		line := make([]note.Event, len(cp))
		for i, p := range cp {
			line[i] = entry[i]
			line[i].Pitch = float64(p)
		}
		return line
	}
	return nil
}

// EpisodeConfig is the setup of an episode
type EpisodeConfig struct {
	// Voices are the voice that plays the fragment of the subject and the voice that plays the fragment of
	// the countersubject with it (default 0 and 1). A voice that is not there rests; with no
	// countersubject the second voice rests
	Voices []int
	// Repeats is the number of times the fragment is played (default 3)
	Repeats int
	// Step is the number of scale degrees the fragment moves at every repeat: -1 (the default) is a
	// sequence down by step, 1 is up by step, 2 up by thirds. 0 is taken as -1
	Step int
	// Shift is the key area at the start, in degrees from the subject (default 0; the octave is chosen
	// to suit the voice)
	Shift int
	// Fragment is the length in steps of the head of the subject that is repeated (default one bar). It is
	// cut at the end of a note, but at least one note is used
	Fragment float64
}

// Episode adds a sequence: a fragment from the head of the subject, with the head of the countersubject
// against it, played again and again, each time Step degrees up or down. Because the countersubject is
// invertible (no fifths, intervals of thirds, sixths and octaves) the two stay consonant in every
// repeat. The episode starts at the next barline. It returns the shift in degrees where the sequence would
// have gone on, which is the key area to aim for with Middle
func (f *Fugue) Episode(c EpisodeConfig) int {
	voices := c.Voices
	if len(voices) == 0 {
		voices = []int{0, 1}
	}
	if voices[0] < 0 || voices[0] >= len(f.Voices) {
		return c.Shift
	}
	repeats := c.Repeats
	if repeats <= 0 {
		repeats = 3
	}
	step := c.Step
	if step == 0 {
		step = -1
	}
	frag := c.Fragment
	if frag <= 0 {
		frag = f.Config.bar()
	}
	count, length := 0, 0.0
	for count == 0 || (count < len(f.Subject.Durations) && length+f.Subject.Durations[count] <= frag+1e-9) {
		length += f.Subject.Durations[count]
		count++
	}
	beat := f.Config.beat()
	period := math.Ceil(length/beat-1e-9) * beat
	start := f.next()
	n := f.Key.Len()

	var degrees []int
	for r := range repeats {
		for _, d := range f.Subject.Degrees[:count] {
			degrees = append(degrees, d+step*r)
		}
	}
	lead := voices[0]
	s0 := fit(f.Key, degrees, c.Shift, f.Config.centers(len(f.Voices))[lead])

	head := Subject{f.Subject.Degrees[:count], f.Subject.Durations[:count]}
	var counter Subject
	partner := -1
	if len(voices) > 1 && voices[1] >= 0 && voices[1] < len(f.Voices) && voices[1] != lead && len(f.Counter.Degrees) > 0 {
		partner = voices[1]
		k := 0
		t := 0.0
		for k < len(f.Counter.Durations) && t < length-1e-9 {
			t += f.Counter.Durations[k]
			k++
		}
		counter = Subject{f.Counter.Degrees[:k], f.Counter.Durations[:k]}
	}
	for r := range repeats {
		t := start + float64(r)*period
		s := s0 + step*r
		f.add(lead, head.Events(f.Key, t, s))
		if partner >= 0 {
			if partner > lead {
				s -= n
			}
			f.add(partner, counter.Events(f.Key, t, s))
		}
	}
	f.section("episode", start)
	return s0 + step*repeats
}

// Middle adds one entry of the subject or the answer in a new key area, at the next barline: shift moves it
// that many degrees (the octave is chosen to suit the voice; 2 is the relative major of a minor key, 4
// the dominant, 3 the subdominant). When withCounter is set the countersubject is played in the voice next to it,
// above the entry if there is a voice above, otherwise below, which is the inversion at the octave. The
// other voices rest
func (f *Fugue) Middle(voice int, kind Kind, shift int, withCounter bool) {
	if voice < 0 || voice >= len(f.Voices) {
		return
	}
	start := f.next()
	base := f.base(kind)
	total := fit(f.Key, base.Degrees, shift, f.Config.centers(len(f.Voices))[voice])
	f.add(voice, base.Events(f.Key, start, total))
	f.Entries = append(f.Entries, Entry{Voice: voice, Start: start, Kind: kind, Shift: total, Section: "middle entry"})
	if withCounter && len(f.Counter.Degrees) > 0 && len(f.Voices) > 1 {
		cv, cs := voice-1, total
		if voice == 0 {
			cv, cs = 1, cs-f.Key.Len()
		}
		f.add(cv, f.counter(kind).Events(f.Key, start, cs))
	}
	f.section("middle entry", start)
}

// Stretto adds entries that overlap: the voices enter at the next barline, one after the other, distance
// steps apart, alternating subject and answer, so the next voice comes in before the one before it has
// finished. voices are the voices in order of entry (nil: the order of the exposition). It
// returns ErrStretto, adding nothing, if the distance is not positive, there are fewer than two voices,
// or two entries form a dissonance on the beats; StrettoDistances finds the distances that work
func (f *Fugue) Stretto(distance float64, voices []int) error {
	if voices == nil {
		voices = f.Config.order(len(f.Voices))
	}
	seen := map[int]bool{}
	for _, v := range voices {
		if v < 0 || v >= len(f.Voices) || seen[v] {
			return ErrStretto
		}
		seen[v] = true
	}
	if !(distance > 0) || len(voices) < 2 {
		return ErrStretto
	}
	start := f.next()
	centers := f.Config.centers(len(f.Voices))
	entries := make([][]note.Event, len(voices))
	var shifts []int
	for i, v := range voices {
		base := f.base(Kind(i % 2))
		s := fit(f.Key, base.Degrees, 0, centers[v])
		shifts = append(shifts, s)
		entries[i] = base.Events(f.Key, start+float64(i)*distance, s)
	}
	if strettoClashes(entries, f.Config.beat()) > 0 {
		return ErrStretto
	}
	for i, v := range voices {
		f.add(v, entries[i])
		f.Entries = append(f.Entries, Entry{Voice: v, Start: start + float64(i)*distance, Kind: Kind(i % 2), Shift: shifts[i], Section: "stretto"})
	}
	f.section("stretto", start)
	return nil
}

// CloseConfig is the setup of the closing
type CloseConfig struct {
	// Bars is the length of the closing (default 4): half of it on the dominant, the rest on the tonic
	Bars int
	// LeadingTone raises the seventh degree of the dominant chord by a semitone when it lies a whole tone
	// below the tonic (the C of D minor becomes C sharp), as in the harmonic minor
	LeadingTone bool
}

// Close adds the ending at the next barline: the lowest voice holds a pedal, first on the dominant and then on
// the tonic, and the other voices hold the notes of the chord nearest to where they were, so the voices
// move as little as possible. The last chord is the tonic triad. It does not check for parallel fifths or
// octaves between the old material and the chords
func (f *Fugue) Close(c CloseConfig) {
	bars := c.Bars
	if bars <= 0 {
		bars = 4
	}
	voices := len(f.Voices)
	n := f.Key.Len()
	fifth := f.fifth()
	bar := f.Config.bar()
	start := f.next()
	centers := f.Config.centers(voices)

	last := make([]float64, voices) // where each voice was
	for v := range voices {
		last[v] = centers[v]
		if l := f.Voices[v]; len(l) > 0 {
			last[v] = l[len(l)-1].Pitch
		}
	}
	bass := voices - 1
	tonic := fit(f.Key, []int{0}, 0, centers[bass])
	dominant := tonic - (n - fifth)

	domBars := bars / 2
	for part, chord := range []struct {
		root int
		len  float64
	}{{fifth, float64(domBars) * bar}, {0, float64(bars-domBars) * bar}} {
		if chord.len == 0 {
			continue
		}
		t := start
		if part == 1 {
			t += float64(domBars) * bar
		}
		pedal := tonic
		if part == 0 {
			pedal = dominant
		}
		below := f.Key.Degree(pedal)
		f.add(bass, []note.Event{{Pitch: below, Start: t, Duration: chord.len}})
		last[bass] = below
		used := map[int]bool{int(math.Round(below)) % 12: true}
		for v := bass - 1; v >= 0; v-- {
			tones := []int{chord.root, chord.root + 2, chord.root + 4}
			if part == 1 && v == 0 {
				tones = tones[:1] // the top voice ends on the tonic
			}
			p, ok := f.chordTone(tones, below, last[v], c.LeadingTone && part == 0, used)
			if !ok {
				continue
			}
			f.add(v, []note.Event{{Pitch: p, Start: t, Duration: chord.len}})
			last[v], below = p, p
			used[int(math.Round(p))%12] = true
		}
	}
	f.section("close", start)
}

// chordTone is the pitch of a degree of tones that lies above floor and nearest to near,
// preferring a pitch class that is not used yet so the chord is complete
func (f *Fugue) chordTone(tones []int, floor, near float64, leading bool, used map[int]bool) (float64, bool) {
	n := f.Key.Len()
	best, bestDist, ok := 0.0, math.Inf(1), false
	for _, d := range tones {
		for m := -6; m <= 6; m++ {
			p := f.Key.Degree(d + m*n)
			if leading && math.Mod(p-f.Key.Root()+120, 12) == 10 {
				p++
			}
			if p <= floor+1e-9 {
				continue
			}
			dist := math.Abs(p - near)
			if used[int(math.Round(p))%12] {
				dist += 5
			}
			if dist < bestDist {
				best, bestDist, ok = p, dist, true
			}
		}
	}
	return best, ok
}
