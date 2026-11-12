package scale

import (
	"math"
	"strings"
	"testing"
)

// FuzzParseSclSane: arbitrary text never panics ParseScl, and a tuning that parses is sane and usable
func FuzzParseSclSane(f *testing.F) {
	for _, s := range []string{
		"! meanquar.scl\n!\nMeantone\n 7\n!\n 76.04900\n 193.15686\n 310.26471\n 5/4\n 503.42157\n 696.57843\n 2/1\n",
		"12 tone\n12\n100.\n200.\n300.\n400.\n500.\n600.\n700.\n800.\n900.\n1000.\n1100.\n1200.\n",
		"\n1\n2\n",
		"d\n2\n3/2\n2\n",
		"d\n1\n1200.0 octave\n",
		"d\n0\n",
		"d\n3\n5/4\n3/2\n",
		"d\n-1\n",
		"d\n1\n0/0\n",
		"d\n1\n1/0\n",
		"d\n1\n1e999\n",
		"d\n1\nNaN.\n",
		"",
		"!\n!\n",
		"d\n99999999999999999999\n",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, text string) {
		s, err := ParseScl(strings.NewReader(text))
		if err != nil {
			return
		}
		if len(s.Cents) == 0 {
			t.Fatal("no pitches without an error")
		}
		prev := 0.0
		for _, c := range s.Cents {
			if !(c > prev) || math.IsInf(c, 0) {
				t.Fatalf("cents %v do not ascend from 0", s.Cents)
			}
			prev = c
		}
		if p := s.Period(); !(p > 0) {
			t.Fatalf("period %v", p)
		}
		_ = s.Steps()
		sc := s.Scale(60)
		_ = sc.Quantize(61.3)
	})
}
