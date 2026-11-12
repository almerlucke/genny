package grammar

import "testing"

// FuzzExpand: a grammar made of arbitrary bytes (possibly recursive, with odd weights) expands without
// a panic and within MaxLen
func FuzzExpand(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, 6)
	f.Add([]byte{255, 0, 0, 0, 0}, 100)
	f.Add([]byte{}, -3)
	f.Add([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 1000)
	f.Fuzz(func(t *testing.T, data []byte, depth int) {
		names := []string{"S", "A", "B", "C"}
		g := New[int]()
		pos := 0
		next := func() byte {
			if pos >= len(data) {
				return 0
			}
			pos++
			return data[pos-1]
		}
		for _, name := range names {
			if pos >= len(data) {
				break
			}
			var alts []Alt[int]
			for range 1 + int(next()%3) {
				var syms []Symbol[int]
				for range int(next() % 4) {
					b := next()
					if b%2 == 0 {
						syms = append(syms, Term(int(b)))
					} else {
						syms = append(syms, Ref[int](names[int(b/2)%len(names)]))
					}
				}
				alts = append(alts, Choice(float64(1+next()%5), syms...))
			}
			g.Rule(name, alts...)
		}
		if g.Check("S") != nil {
			return
		}
		out := g.Expand("S", min(depth, 60))
		if len(out) > MaxLen {
			t.Fatalf("%d symbols", len(out))
		}
		p := NewPlayer(g, "S", min(depth, 60))
		for range 20 {
			p.Generate()
		}
	})
}
