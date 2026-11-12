package sieve

import (
	"slices"
	"testing"
)

// FuzzSieveAlgebra: any parameters build sieves that obey the laws of sets and never panic
func FuzzSieveAlgebra(f *testing.F) {
	f.Add(3, 1, 0, 20, 4, 0, -5, 25)
	f.Add(1, 0, 5, 4, 2, 7, 0, 0)
	f.Add(-0, -7, -100, 100, 12, -3, -50, 50)
	f.Fuzz(func(t *testing.T, m1, s1, lo1, hi1, m2, s2, lo2, hi2 int) {
		clip := func(lo, hi int) (int, int) {
			lo = min(max(lo, -5000), 5000)
			hi = min(max(hi, -5000), 5000)
			return lo, hi
		}
		lo1, hi1 = clip(lo1, hi1)
		lo2, hi2 = clip(lo2, hi2)
		if m1 < 1 || m2 < 1 {
			return // New panics by contract
		}
		a, b := New(m1, s1, lo1, hi1), New(m2, s2, lo2, hi2)
		for _, s := range []Sieve{a, b} {
			if !slices.IsSorted(s.Values()) || s.Len() != len(s.Values()) {
				t.Fatal("values not sorted")
			}
		}
		u, i := a.Union(b), a.Intersection(b)
		for _, x := range u.Values() {
			if !a.Contains(x) && !b.Contains(x) {
				t.Fatalf("%d in the union only", x)
			}
		}
		for _, x := range i.Values() {
			if !a.Contains(x) || !b.Contains(x) {
				t.Fatalf("%d in the intersection only", x)
			}
		}
		if u.Len() < max(a.Len(), b.Len()) || i.Len() > min(a.Len(), b.Len()) {
			t.Fatal("sizes")
		}
		d := a.Difference(b)
		for _, x := range d.Values() {
			if !a.Contains(x) || b.Contains(x) {
				t.Fatalf("%d wrongly in the difference", x)
			}
		}
		if a.Difference(b).Len()+i.Len() != a.Len() {
			t.Fatal("difference + intersection != a")
		}
		c := a.Complement()
		lo, hi := a.Range()
		if hi >= lo && c.Len()+a.Len() != hi-lo+1 {
			t.Fatalf("complement %d + %d != range", c.Len(), a.Len())
		}
		for _, x := range c.Values() {
			if a.Contains(x) {
				t.Fatal("complement overlaps")
			}
		}
		// generators never panic and stay in the sieve
		g := NewGen(a, false)
		for range a.Len() + 3 {
			v := g.Generate()
			if a.Len() > 0 && !a.Contains(v) {
				t.Fatalf("generated %d, not a member", v)
			}
		}
	})
}
