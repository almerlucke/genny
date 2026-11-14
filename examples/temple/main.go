// temple is a piece built from the measurements of Solomon's temple (1 Kings 6-8, 2 Chronicles 3-4). Every
// dimension in cubits becomes an interval, a number of bars or a rhythm, and the pitches are tuned in just
// intonation (as ratios of a harmonic series on D, through package microtonal), so a measure is literally a
// proportion you can hear.
//
// The piece walks the way a priest walked, from the court through the porch and the holy place into the
// most holy place, and ends with the cloud of glory. The porch and the most holy place last as many bars as they are cubits long, the holy place half a bar a cubit:
//
//	The court (12 bars, the 12 oxen)      altar 20 x 20 x 10, the molten sea, ten lavers
//	The porch (10 bars)                   10 cubits deep
//	The holy place (20 bars)              40 cubits long, 20 wide, 30 high: half a bar a cubit
//	The most holy place (20 bars)         a cube of 20 x 20 x 20
//	The glory (7 bars)                    the cloud fills the house (1 Kings 8:10-11)
//
// How the measures sound:
//
//   - the court: the sea (10 across, 30 around, 5 high, standing on 12 oxen, three facing each way) is a cycle of
//     12 bell strokes, one a bar, in four groups of three. The sea is 5 : 10 : 30 = 1 : 2 : 6, so the three oxen
//     of a group sound the partials 2, 4 and 12, and the four directions stand on 1, 4/3, 3/2 and 2 (unison,
//     fourth, fifth and octave). The ten lavers (five on each side) climb and descend the partials 4 to 8 against
//     the four beats of the bar. The offerings at the dedication (22,000 oxen and 120,000 sheep) are 22 drum
//     strokes against 120 ticks, a polyrhythm of 11 : 60
//
//   - the porch is an isorhythm, a medieval form with a rhythm cycle (the talea) and a pitch cycle (the color) of
//     different lengths. The talea is the capital's height, the lily work and the chains: 5 + 4 + 7 = 16 steps, one
//     bar. The color is seven measures: 4 (lily work), 5 (capital), 7 (chains), 10 (depth), 12 and 18 (the pillar's
//     circumference and height) and 20 (width), each reduced to one of the four notes of the theme of the holy
//     place (D E F# A: the measure modulo 4), in the octave its size asks. Jachin plays the color forward,
//     Boaz answers backward, a step later and two octaves lower
//
//   - the holy place is a proportion canon. It is 40 long, 20 wide and 30 high, and three voices play the same theme
//     in those proportions: the note lengths of the voice of the length are multiplied by 4, those of the voice of
//     the height by 3 and those of the voice of the width by 2 (the longer the room, the slower). They start
//     together and drift apart, and meet again after 120 steps (the least common multiple of 40, 20 and 30), where
//     the chord of the three storeys of side chambers (5 : 6 : 7) sounds over a pedal that moves between 1, 3/2 and 4/3
//
//   - the most holy place is a cube, 20 x 20 x 20, so everything in it is a mirror. In time it is a palindrome: the
//     second ten bars are the first ten backwards. In pitch it is a mirror around D: the cherub's wing above goes
//     down the overtones 4, 3, 8/3 and 5/2 to the center and the wing below goes up the same intervals mirrored
//     (the undertones), so the harmony is major and minor at once. The two wings that touch in the middle start
//     a few cents apart and come together into a pure unison exactly in the middle, then part again. The room is
//     overlaid with gold: a high bell at each wing and a higher one in the middle
//
//   - the sea and the lavers of the court return in the most holy place: the 12 strokes of the sea go round forwards and
//     then backwards, and the lavers climb and descend, the second ten bars being the first ten backwards
//
//   - the offerings of the court return, the same 22 strokes against 120 ticks (11 : 60), slower: over 10 bars under the
//     second half of the holy place and then over the 20 bars of the most holy place (two and a half times slower than in the
//     court), the strokes swelling to the middle of the cube and fading again, which keeps the cube from standing still
//
//   - light in the most holy place: the ten lampstands (five on each side), which stood before the inner sanctuary, are lit
//     one by one, one every two bars, high on the partials 8 to 17 on two harps, and flicker on their own periods. The
//     light only grows, into the glory, so unlike the rest of the cube it is not a mirror
//
//   - a gentle melody on a piano through the first four rooms. Its rhythm is a cell of 3 + 2 + 2 steps (the seven chains) that goes
//     against the bar of 16 steps and meets it again only every 7 bars; a phrase is four cells and ends on a rest. It walks on
//     the pentatonic scale on D and its register follows the building: low in the court, climbing to the lamps in the most
//     holy place; it stops on a D where the glory begins, which is left to the cloud
//
//   - the glory: the harmonic series 1 to 14 on all fourteen instruments at once (fourteen: the two weeks of
//     the feast of dedication, 1 Kings 8:65), which then fades
//
//     go run ./examples/temple -mid temple.mid
//
// The file needs a synth that follows pitch bend per channel (see package microtonal): each instrument has a
// channel of its own and its own bends
package main

import (
	"flag"
	"fmt"
	"log"
	"maps"
	"math"
	"slices"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/microtonal"
	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/note"
)

const (
	stepsPerBar = 16
	fund        = 38.0 // D2: the fundamental of the harmonic series everything is tuned to
	tempo       = 66

	courtBars  = 12 // the 12 oxen
	porchBars  = 10
	holyBars   = 20 // 40 cubits, at half a bar a cubit
	holiestBar = 20
	gloryBars  = 7 // 7 years of building
)

// bar starts, in steps
var (
	courtStart   = 0.0
	porchStart   = float64(courtBars * stepsPerBar)
	holyStart    = porchStart + porchBars*stepsPerBar
	holiestStart = holyStart + holyBars*stepsPerBar
	gloryStart   = holiestStart + holiestBar*stepsPerBar
	endStep      = gloryStart + gloryBars*stepsPerBar
)

// partial is the pitch of a ratio above the fundamental
func partial(ratio float64) float64 { return fund + 12*math.Log2(ratio) }

// voice is an instrument: all its notes go to one channel with its own pitch bends
type voice struct {
	name    string
	channel int
	program int
	pan     int
	events  []note.Event
	cc      []midi.Controller
}

var voices = map[int]*voice{}

func def(channel int, name string, program, pan int) {
	voices[channel] = &voice{name: name, channel: channel, program: program, pan: pan}
}

func (v *voice) add(pitch, vel, start, dur float64) {
	v.events = append(v.events, note.Event{Pitch: pitch, Velocity: math.Min(math.Max(vel, 0.05), 1), Start: start, Duration: dur})
}

// monophonic makes the notes of a voice follow each other: notes at the same time are moved a little apart and a
// note ends where the next begins, because one channel has one pitch bend
func monophonic(events []note.Event) []note.Event {
	slices.SortStableFunc(events, func(a, b note.Event) int {
		switch {
		case a.Start < b.Start:
			return -1
		case a.Start > b.Start:
			return 1
		}
		return 0
	})
	for i := 1; i < len(events); i++ {
		if events[i].Start <= events[i-1].Start {
			events[i].Start = events[i-1].Start + 0.25
		}
	}
	for i := 0; i+1 < len(events); i++ {
		if end := events[i].Start + events[i].Duration; end > events[i+1].Start-0.05 {
			events[i].Duration = events[i+1].Start - 0.05 - events[i].Start
		}
	}
	return events
}

func main() {
	seed := flag.Uint64("seed", 1, "seed of the random stream (the offering ticks use it)")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	def(0, "molten sea", 14, 64)  // tubular bells
	def(1, "lavers left", 11, 24) // vibraphone
	def(2, "lavers right", 9, 104)
	def(3, "jachin", 19, 98) // church organ
	def(4, "boaz", 19, 30)
	def(5, "gold bells", 11, 64) // vibraphone
	def(6, "lamps left", 46, 22) // harp
	def(7, "lamps right", 46, 106)
	def(8, "side chambers 5", 52, 44) // choir aahs
	def(10, "side chambers 6", 52, 64)
	def(11, "side chambers 7", 52, 84)
	def(12, "altar and room", 89, 64) // warm pad
	def(13, "wing four", 52, 64)
	def(14, "glory", 91, 64)
	def(15, "melody", 0, 64) // piano

	// ---- the court: the altar, the sea on twelve oxen, the ten lavers, the offerings
	altar := voices[12] // 20 x 20 x 10 = 2 : 2 : 1, a low D under everything
	altar.add(partial(1), 0.45, courtStart, float64(courtBars*stepsPerBar)+4)

	sea := voices[0]
	directions := []float64{1, 4.0 / 3, 3.0 / 2, 2} // unison, fourth, fifth and octave
	oxen := []float64{2, 4, 12}                     // the sea is 5 : 10 : 30 = 1 : 2 : 6
	ox := func(i int) (pitch, vel float64) {        // the pitch and the strength of the stroke of ox i
		d, k := i/3, i%3
		vel = 0.6
		if k == 0 {
			vel = 0.78
		}
		return partial(directions[d] * oxen[k]), vel
	}
	for i := 0; i < 12; i++ {
		pitch, vel := ox(i)
		sea.add(pitch, vel, courtStart+float64(i*stepsPerBar), stepsPerBar-0.5)
	}

	left, right := 0, 0
	for t := 4 * stepsPerBar; t < courtBars*stepsPerBar; t += 2 {
		switch t % 4 {
		case 0: // the five lavers on the left climb the partials 4..8
			voices[1].add(partial(float64(4+left%5)), 0.4, float64(t), 3.5)
			left++
		case 2: // those on the right descend them
			voices[2].add(partial(float64(8-right%5)), 0.4, float64(t), 3.5)
			right++
		}
	}

	// the offerings of the dedication, 22 oxen against 120 sheep over eight bars; the phrase comes back in the second half of
	// the holy place and in the most holy place (below)
	var drums []note.Event
	offering := func(start, length, shift float64, arc func(t float64) float64) { // shift: where the first stroke falls, in strokes (0.5 centers the pattern)
		for i := 0; i < 22; i++ {
			t := start + (float64(i)+shift)*length/22
			drums = append(drums, note.Event{Pitch: 41, Velocity: arc(t), Start: t, Duration: 1})
		}
		for i := 0; i < 120; i++ {
			t := start + (float64(i)+shift)*length/120
			drums = append(drums, note.Event{Pitch: 37, Velocity: 0.14 + 0.1*genny.Float64(), Start: t, Duration: 0.5})
		}
	}
	offering(float64(4*stepsPerBar), float64(8*stepsPerBar), 0, func(t float64) float64 { return 0.35 + 0.25*(t-4*stepsPerBar)/(8*stepsPerBar) })

	// ---- the porch: an isorhythm. Talea 5 + 4 + 7 (capital, lily work, chains), color 4 5 7 10 12 18 20 (from the measures)
	talea := []float64{5, 4, 7}
	color := []float64{4, 5, 7, 10, 12, 18, 20}
	notes := []float64{1, 9.0 / 8, 5.0 / 4, 3.0 / 2} // D E F# A, the notes of the theme of the holy place
	pitch := func(m float64) float64 {               // a measure as a note: its place modulo 4, an octave higher for the larger measures
		r := notes[int(m)%4] * 2
		if m >= 10 {
			r *= 2
		}
		return r
	}
	jachin, boaz := voices[3], voices[4]
	for i, t := 0, porchStart; t < holyStart; i++ {
		d := talea[i%3]
		jachin.add(partial(pitch(color[i%7])), 0.3+0.35*(t-porchStart)/(porchBars*stepsPerBar), t, math.Min(d, holyStart-t)-0.2)
		t += d
	}
	for i, t := 0, porchStart+stepsPerBar/2; t < holyStart; i++ { // the left pillar answers, backwards, half a bar later
		d := talea[(i+1)%3]
		boaz.add(partial(pitch(color[6-i%7])/2), 0.3+0.3*(t-porchStart)/(porchBars*stepsPerBar), t, math.Min(d, holyStart-t)-0.2)
		t += d
	}

	// ---- the holy place: a proportion canon, durations times 4 (length), 3 (height) and 2 (width). It has two halves of
	// 20 cubits (10 bars): in the second half all durations are divided by 1.5, the canon starts anew at one and a half times the speed
	theme := []float64{4, 4.5, 5, 6, 4.5, 5, 6, 8} // partials: D E F# A E F# A D
	lengths := []float64{2, 1, 1, 2, 1, 1, 1, 1}   // 10 steps
	split := holyStart + float64(holyBars/2*stepsPerBar)
	canon := func(from, to, speed float64) {
		for _, v := range []struct {
			ch            int
			scale, octave float64
		}{{8, 4, 0.5}, {3, 3, 1}, {6, 2, 2}} {
			for t := from; t < to; {
				for k, r := range theme {
					d := lengths[k] * v.scale / speed
					if t >= to {
						break
					}
					vel := 0.4
					if k == 0 {
						vel = 0.55
					}
					voices[v.ch].add(partial(r*v.octave), vel, t, math.Min(d, to-t)-0.3)
					t += d
				}
			}
		}
	}
	canon(holyStart, split, 1)
	canon(split, holiestStart, 1.5)
	// whenever the three meet (every 120 steps, in the second half every 60): the chord of the side chambers 5 : 6 : 7 over a pedal
	meet := func(t0, span, r float64) {
		voices[4].add(partial(r), 0.5, t0, span-1)
		for i, ch := range []int{10, 11, 13} {
			voices[ch].add(partial(r*[]float64{2.5, 3, 3.5}[i]), 0.42, t0+float64(i)*3, math.Min(40, span)-float64(i)*3)
		}
	}
	meet(holyStart, 120, 1)
	meet(holyStart+120, split-holyStart-120, 3.0/2)
	meet(split, 80, 4.0/3) // at 1.5 times the speed the voices meet every 80 steps
	meet(split+80, 80, 3.0/2)
	voices[4].add(partial(1), 0.5, split+120, holiestStart-split-120-1) // home on the last 40
	voices[5].add(partial(16), 0.4, split, 12)                          // a bell where the second half begins

	// the offerings return, slower: 10 bars under the second half of the holy place (the court took 8) and then, in the most holy place,
	// a single round over all 20 bars, two and a half times slower than in the court, centered so that it is a mirror, the strokes
	// swelling to the middle of the cube and fading again
	center := holiestStart + float64(holiestBar/2*stepsPerBar)
	arc := func(t float64) float64 { return 0.25 + 0.4*(1-math.Abs(t-center)/(center-split)) }
	offering(split, holiestStart-split, 0, arc)
	offering(holiestStart, gloryStart-holiestStart, 0.5, arc)

	// ---- the most holy place: a cube, so a mirror in time and in pitch
	cherubim := holiestStart
	bar := float64(stepsPerBar)
	upper := []float64{4, 3, 8.0 / 3, 5.0 / 2, 2}       // the wing above comes down to the middle
	lower := []float64{1, 4.0 / 3, 3.0 / 2, 8.0 / 5, 2} // the one below goes up, mirrored
	for i := 0; i < 10; i++ {
		k := min(i, 9-i) // the second ten bars are the first backwards
		voices[8].add(partial(upper[k]), 0.5, cherubim+float64(2*i)*bar, 2*bar-0.4)
		voices[13].add(partial(lower[k]), 0.5, cherubim+float64(2*i)*bar, 2*bar-0.4)
	}
	for i, cents := range []float64{30, 20, 12, 5, 0, 0, 5, 12, 20, 30} { // the wings that touch in the middle
		t := cherubim + float64(5+i)*bar
		voices[10].add(partial(2)+cents/100, 0.45, t, bar-0.3)
		voices[11].add(partial(2)-cents/100, 0.45, t, bar-0.3)
	}
	for _, b := range []float64{0, 5, 14, 19} { // the gold
		voices[5].add(partial(8), 0.4, cherubim+b*bar, 12)
	}
	voices[5].add(partial(16), 0.45, cherubim+10*bar-2, 14)
	// the sea and the lavers of the court come back: the 12 strokes of the sea go round forwards and then backwards, 24 strokes
	// in the 20 bars (a mirror in time as well), and the lavers climb and descend as in the court, the second ten bars
	// being the first ten backwards
	for j := 0; j < 24; j++ {
		i := j
		if j >= 12 {
			i = 23 - j
		}
		pitch, vel := ox(i)
		voices[0].add(pitch, vel*0.85, cherubim+(float64(j)+0.5)*holiestBar*stepsPerBar/24, holiestBar*stepsPerBar/24-0.5)
	}
	left, right = 0, 0
	for t := 0.0; t < holiestBar*stepsPerBar/2; t += 2 {
		mirror := holiestBar*stepsPerBar - 4 - t // the same note, as many steps from the end as it is from the start
		switch int(t) % 4 {
		case 0:
			p := partial(float64(4 + left%5))
			voices[1].add(p, 0.32, cherubim+t, 3.5)
			voices[1].add(p, 0.32, cherubim+mirror, 3.5)
			left++
		case 2:
			p := partial(float64(8 - right%5))
			voices[2].add(p, 0.32, cherubim+t, 3.5)
			if mirror != t { // the middle note is its own mirror
				voices[2].add(p, 0.32, cherubim+mirror, 3.5)
			}
			right++
		}
	}

	// light: the ten lampstands (five each side, 1 Kings 7:49) are lit one by one, one every two bars, high on the partials 8 to 17,
	// and flicker on their own period. They are not part of the mirror: the light only grows, into the glory
	for i := 0; i < 10; i++ {
		lit := cherubim + float64(2*i)*bar
		lamp := voices[6+i%2]
		lamp.add(partial(float64(8+i)), 0.55, lit, 10)
		for t, period := lit+float64(8+i), float64(8+i); t < gloryStart-4; t += period {
			if genny.Float64() < 0.8 { // the flame flickers
				lamp.add(partial(float64(8+i)), 0.2+0.2*genny.Float64(), t, 6)
			}
		}
	}
	room := voices[12] // the room, 20 beats to a breath
	room.add(partial(2), 0.35, cherubim, holiestBar*stepsPerBar)
	for t := 0.0; t < holiestBar*stepsPerBar; t += 8 {
		room.cc = append(room.cc, midi.Controller{Step: cherubim + t, Number: 11, Value: int(88 + 38*math.Sin(2*math.Pi*t/(20*4)))})
	}

	// ---- a gentle melody on a piano through the first four rooms. Its rhythm is a cell of 3 + 2 + 2 steps (the seven chains), which goes
	// against the bar of 16 steps and only falls on the bar again every 7 bars; a phrase is four cells and ends on a rest. Its notes
	// are the pentatonic scale on D (D E F# A B) in just intonation, and its register follows the building: it starts low in the court,
	// climbs through the porch and the holy place to the lamps in the most holy place, and stops on a D where the glory begins
	pent := []float64{1, 9.0 / 8, 5.0 / 4, 3.0 / 2, 5.0 / 3}
	degree := func(i int) float64 { // the ratio above the fundamental of step i of the scale, counted from D2
		oct, d := i/5, i%5
		if d < 0 {
			d += 5
			oct--
		}
		return pent[d] * math.Pow(2, float64(oct))
	}
	keys := [][2]float64{{0, 8}, {12, 9}, {22, 11}, {42, 13}, {62, 15}} // the bar and the step the melody is drawn to there
	target := func(t float64) float64 {
		b := t / stepsPerBar
		for k := 1; k < len(keys); k++ {
			if b <= keys[k][0] {
				f := (b - keys[k-1][0]) / (keys[k][0] - keys[k-1][0])
				return keys[k-1][1] + f*(keys[k][1]-keys[k-1][1])
			}
		}
		return keys[len(keys)-1][1]
	}
	loudness := func(t float64) float64 { // gentle everywhere, a little more in the holy place and softer in the glory
		switch {
		case t < porchStart:
			return 0.8
		case t < holyStart:
			return 0.9
		case t < holiestStart:
			return 1
		case t < gloryStart:
			return 0.85
		}
		return 0.8
	}
	melody := voices[15]
	cell := []float64{3, 2, 2}
	cur := 8
	last := gloryStart - 24 // the melody stops where the glory begins
	for t, c := 0.0, 0; t < last; c++ {
		for k, d := range cell {
			if t >= last {
				break
			}
			tg := target(t)
			rest := (c%4 == 3 && k == 2) || (k > 0 && genny.Float64() < 0.12) // the phrase breathes
			if k == 0 && c%4 == 0 {                                           // a phrase begins on a note of the chord D F# A, as near to the target as possible
				best := cur
				for i := int(tg) - 3; i <= int(tg)+3; i++ {
					if dd := ((i % 5) + 5) % 5; (dd == 0 || dd == 2 || dd == 3) && math.Abs(float64(i)-tg) < math.Abs(float64(best)-tg) {
						best = i
					}
				}
				cur = best
			} else {
				var weights [5]float64
				total := 0.0
				for j, delta := range []int{-2, -1, 0, 1, 2} {
					w := math.Exp(-math.Pow(float64(cur+delta)-tg, 2) / 8)
					switch delta {
					case 0:
						w *= 0.3
					case -2, 2:
						w *= 0.5
					}
					weights[j] = w
					total += w
				}
				x := genny.Float64() * total
				for j, w := range weights {
					if x < w || j == 4 {
						cur += j - 2
						break
					}
					x -= w
				}
			}
			if !rest {
				vel := (0.5 + 0.12*genny.Float64()) * loudness(t)
				if k == 0 {
					vel += 0.08
				}
				melody.add(partial(degree(cur)), vel, t, d)
			}
			t += d
		}
	}
	melody.add(partial(degree(15)), 0.5, last, 24) // a last D, as the glory begins

	// ---- the glory: the cloud fills the house, the series 1..14 on fourteen instruments
	cloud := []int{3, 4, 8, 10, 11, 13, 12, 0, 5, 1, 2, 6, 7, 14}
	for i, ch := range cloud {
		start := gloryStart + float64(i)*5
		voices[ch].cc = append(voices[ch].cc, midi.Controller{Step: gloryStart - 1, Number: 11, Value: 110})
		voices[ch].add(partial(float64(i+1)), 0.55-0.2*float64(i)/14, start, endStep-start+6)
		for k := 0; k < 12; k++ { // the cloud thins out over the last bars
			voices[ch].cc = append(voices[ch].cc, midi.Controller{Step: endStep - 40 + float64(k)*4, Number: 11, Value: int(110 * (1 - float64(k+1)/12))})
		}
	}

	// ---- the tracks
	var tracks []midi.Track
	for _, ch := range slices.Sorted(maps.Keys(voices)) {
		v := voices[ch]
		if len(v.events) == 0 {
			continue
		}
		rendered := microtonal.Render(monophonic(v.events), nil, microtonal.Strategy{Channel: ch, Program: v.program, Name: v.name})
		for _, tr := range rendered {
			tr.Controllers = append(tr.Controllers, midi.Controller{Step: 0, Number: 10, Value: v.pan}, midi.Controller{Step: 0, Number: 11, Value: 110})
			tr.Controllers = append(tr.Controllers, v.cc...)
			tracks = append(tracks, tr)
		}
	}
	tracks = append(tracks, midi.Track{Name: "offerings", Channel: 9, Events: drums})

	fmt.Printf("%-22s bars %2d - %2d  %s\n", "the court", 1, courtBars, "altar, molten sea on 12 oxen, ten lavers")
	fmt.Printf("%-22s bars %2d - %2d  %s\n", "the porch", courtBars+1, courtBars+porchBars, "isorhythm: talea 5+4+7, color 4 5 7 10 12 18 20")
	fmt.Printf("%-22s bars %2d - %2d  %s\n", "the holy place", courtBars+porchBars+1, courtBars+porchBars+holyBars, "canon 4:2:3, chambers 5:6:7")
	fmt.Printf("%-22s bars %2d - %2d  %s\n", "the most holy place", int(holiestStart)/stepsPerBar+1, int(gloryStart)/stepsPerBar, "a cube: a mirror in time and pitch")
	fmt.Printf("%-22s bars %2d - %2d  %s\n", "the glory", int(gloryStart)/stepsPerBar+1, int(endStep)/stepsPerBar, "partials 1..14")
	n := 0
	for _, t := range tracks {
		n += len(t.Events)
	}
	fmt.Printf("%d notes on %d tracks, about %.0f seconds at %d bpm\n", n, len(tracks), endStep/4*60/tempo, tempo)

	if *midiFile != "" {
		if err := midi.WriteFile(*midiFile, midi.File{Tempo: tempo, StepsPerBeat: 4, Tracks: tracks}); err != nil {
			log.Fatal(err)
		}
		fmt.Println("written to", *midiFile)
	}
}
