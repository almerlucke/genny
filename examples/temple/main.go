// temple is a piece built from the measurements of Solomon's temple (1 Kings 6-8, 2 Chronicles 3-4). Every
// dimension in cubits becomes an interval, a number of bars or a rhythm, and the pitches are tuned in just
// intonation (as ratios of a harmonic series on D, through package microtonal), so a measure is literally a
// proportion you can hear.
//
// The piece walks the way a priest walked, from the court through the porch and the holy place into the
// most holy place, and ends with the cloud of glory. The bars of the three rooms are their lengths in cubits:
//
//	The court (12 bars, the 12 oxen)      altar 20 x 20 x 10, the molten sea, ten lavers
//	The porch (10 bars)                   10 cubits deep
//	The holy place (40 bars)              40 cubits long, 20 wide, 30 high
//	The most holy place (20 bars)         a cube of 20 x 20 x 20
//	The glory (7 bars)                    the cloud fills the house (1 Kings 8:10-11)
//
// How the measures sound:
//
//   - the sea (10 across, 30 around, 5 high, standing on 12 oxen, three facing each way): a cycle of 12 bell
//     strokes, one a bar, in four groups of three. The sea is 5 : 10 : 30 = 1 : 2 : 6, so the three oxen of a
//     group sound the partials 2, 4 and 12, and the four directions stand on 1, 4/3, 3/2 and 2 (unison,
//     fourth, fifth and octave)
//
//   - the ten lavers (five on each side): five notes on the left climb the partials 4 5 6 7 8 and five on the
//     right descend them, against the four beats of the bar
//
//   - the offerings at the dedication (22,000 oxen and 120,000 sheep, 1 Kings 8:63): 22 drum strokes against
//     120 ticks over the same eight bars, a polyrhythm of 22 : 120 = 11 : 60
//
//   - the pillars Jachin and Boaz (18 high, 12 around, so 3 : 2): Jachin sounds the third partial, Boaz the
//     second (a fifth), and Jachin is struck every 18 beats and Boaz every 12, a 3 : 2 rhythm that comes
//     together every 36 beats
//
//   - the capitals' two rows of pomegranates (200 in all): 200 bells, rows of 100 alternating, each row
//     walking on the partials 8 to 24
//
//   - the holy place (40 : 20 : 30 = 4 : 2 : 3): the room is the interval 2 : 3 : 4, an octave built from a fifth
//     and a fourth. The three storeys of side chambers (5, 6 and 7 cubits wide, each 5 high) are the chord
//     5 : 6 : 7 (a major third, a minor third and a seventh), which changes its root every 5 bars, the height of
//     a storey. The ten lampstands (five on each side) are lit one by one, one every 4 bars, each on a partial from
//     8 to 17 that flickers on its own period
//
//   - the most holy place (20 x 20 x 20): a cube, so unison. The two cherubim, each 10 high (half the room, an
//     octave) with wings of 5 cubits, touch the walls and each other: four wings of 5 make 20, so they enter
//     5 bars apart. The two wings that meet in the middle start apart (a few cents) and come together to a pure
//     unison exactly when the other arrives. The room is overlaid with gold: a high bell at each wing
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
	holyBars   = 40
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
	seed := flag.Uint64("seed", 1, "seed of the random stream (the pomegranates and the flames of the lamps use it)")
	midiFile := flag.String("mid", "", "write the piece to this MIDI file")
	flag.Parse()
	genny.Seed(*seed)

	def(0, "molten sea", 14, 64)  // tubular bells
	def(1, "lavers left", 11, 24) // vibraphone
	def(2, "lavers right", 9, 104)
	def(3, "jachin", 19, 98) // church organ
	def(4, "boaz", 19, 30)
	def(5, "pomegranates", 8, 64) // celesta
	def(6, "lamps left", 46, 22)  // harp
	def(7, "lamps right", 46, 106)
	def(8, "side chambers 5", 52, 44) // choir aahs
	def(10, "side chambers 6", 52, 64)
	def(11, "side chambers 7", 52, 84)
	def(12, "altar and room", 89, 64) // warm pad
	def(13, "wing four", 52, 64)
	def(14, "glory", 91, 64)

	// ---- the court: the altar, the sea on twelve oxen, the ten lavers, the offerings
	altar := voices[12] // 20 x 20 x 10 = 2 : 2 : 1, a low D under everything
	altar.add(partial(1), 0.45, courtStart, float64(courtBars*stepsPerBar)+4)

	sea := voices[0]
	directions := []float64{1, 4.0 / 3, 3.0 / 2, 2} // unison, fourth, fifth and octave
	oxen := []float64{2, 4, 12}                     // the sea is 5 : 10 : 30 = 1 : 2 : 6
	for i := 0; i < 12; i++ {
		d, k := i/3, i%3
		vel := 0.6
		if k == 0 {
			vel = 0.78
		}
		sea.add(partial(directions[d]*oxen[k]), vel, courtStart+float64(i*stepsPerBar), stepsPerBar-0.5)
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

	// the offerings of the dedication, 22 oxen against 120 sheep over eight bars
	var drums []note.Event
	offerStart, offerLen := float64(4*stepsPerBar), float64(8*stepsPerBar)
	for i := 0; i < 22; i++ {
		drums = append(drums, note.Event{Pitch: 41, Velocity: 0.35 + 0.25*float64(i)/22, Start: offerStart + float64(i)*offerLen/22, Duration: 1})
	}
	for i := 0; i < 120; i++ {
		drums = append(drums, note.Event{Pitch: 37, Velocity: 0.14 + 0.1*genny.Float64(), Start: offerStart + float64(i)*offerLen/120, Duration: 0.5})
	}

	// ---- the porch: the pillars Jachin and Boaz, 18 high and 12 around, and the 200 pomegranates
	jachin, boaz := voices[3], voices[4]
	for s := porchStart; s < holyStart; s += 18 * 4 {
		jachin.add(partial(3), 0.6, s, math.Min(18*4, holyStart-s))
	}
	for s := porchStart; s < holyStart; s += 12 * 4 {
		boaz.add(partial(2), 0.55, s, math.Min(12*4, holyStart-s))
	}
	rows := []float64{8, 9, 10, 12, 14, 15, 16, 18, 20, 24}
	idx := [2]int{0, 9}
	span := float64(porchBars * stepsPerBar)
	for i := 0; i < 200; i++ {
		row := i % 2 // two rows of 100
		idx[row] = min(max(idx[row]+genny.Intn(3)-1, 0), len(rows)-1)
		swell := math.Sin(math.Pi * float64(i) / 200)
		voices[5].add(partial(rows[idx[row]]), 0.2+0.3*swell, porchStart+float64(i)*span/200, 0.75)
	}

	// ---- the holy place: the room 2 : 3 : 4, the side chambers 5 : 6 : 7 and the ten lampstands
	roots := []float64{1, 3.0 / 2, 4.0 / 3, 2, 3.0 / 2, 4.0 / 3, 3.0 / 2, 1} // a block of 5 bars, the height of a storey
	rootAt := func(t float64) float64 {
		return roots[min(int((t-holyStart)/(5*stepsPerBar)), len(roots)-1)]
	}
	for j, r := range roots {
		t0 := holyStart + float64(j*5*stepsPerBar)
		jachin.add(partial(r), 0.55, t0, 5*stepsPerBar-1)
		for i, ch := range []int{8, 10, 11} {
			voices[ch].add(partial(r*[]float64{2.5, 3, 3.5}[i]), 0.42, t0+float64(i)*3, 5*stepsPerBar-1-float64(i)*3)
		}
	}
	for i := 0; i < 10; i++ {
		lit := holyStart + float64(i*4*stepsPerBar) // the lamp is lit at bar 4i
		side := voices[6+i%2]
		pitch := func(t float64) float64 {
			r := float64(8+i) * rootAt(t)
			for r > 24 {
				r /= 2
			}
			return partial(r)
		}
		side.add(pitch(lit), 0.6, lit, 8)
		period := float64(8 + i)
		for t := lit + period; t < holiestStart-4; t += period {
			if genny.Float64() < 0.8 { // the flame flickers
				side.add(pitch(t), 0.22+0.2*genny.Float64(), t, 6)
			}
		}
	}

	// ---- the most holy place: a cube, so unison; the wings of the cherubim, 5 cubits each
	cherubim := holiestStart
	voices[8].add(partial(1), 0.5, cherubim, holiestBar*stepsPerBar) // wing 1 touches the wall
	wing2 := voices[10]                                              // wing 2 reaches for the middle
	for i, cents := range []float64{35, 22, 12, 5, 2} {
		wing2.add(partial(2)+cents/100, 0.5, cherubim+float64((5+i)*stepsPerBar), stepsPerBar-0.3)
	}
	wing2.add(partial(2), 0.5, cherubim+10*stepsPerBar, 10*stepsPerBar) // the wings touch: a pure unison
	voices[11].add(partial(2), 0.5, cherubim+10*stepsPerBar, 10*stepsPerBar)
	voices[13].add(partial(1), 0.5, cherubim+15*stepsPerBar, 5*stepsPerBar) // wing 4 touches the other wall
	for w := 0; w < 4; w++ {                                                // the gold
		voices[5].add(partial(16), 0.4, cherubim+float64(w*5*stepsPerBar), 12)
	}
	room := voices[12] // the room, 20 beats to a breath
	room.add(partial(2), 0.35, cherubim, holiestBar*stepsPerBar)
	for t := 0.0; t < holiestBar*stepsPerBar; t += 8 {
		room.cc = append(room.cc, midi.Controller{Step: cherubim + t, Number: 11, Value: int(88 + 38*math.Sin(2*math.Pi*t/(20*4)))})
	}

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
	fmt.Printf("%-22s bars %2d - %2d  %s\n", "the porch", courtBars+1, courtBars+porchBars, "Jachin 3 : Boaz 2, 200 pomegranates")
	fmt.Printf("%-22s bars %2d - %2d  %s\n", "the holy place", courtBars+porchBars+1, courtBars+porchBars+holyBars, "room 2:3:4, chambers 5:6:7, ten lamps")
	fmt.Printf("%-22s bars %2d - %2d  %s\n", "the most holy place", int(holiestStart)/stepsPerBar+1, int(gloryStart)/stepsPerBar, "a cube: unison, four wings of 5")
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
