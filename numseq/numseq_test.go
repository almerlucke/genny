package numseq

import (
	"slices"
	"testing"
)

func take(g interface{ Generate() int }, n int) (out []int) {
	for range n {
		out = append(out, g.Generate())
	}
	return
}

func TestFibonacci(t *testing.T) {
	f := NewFibonacci(10)
	got := take(f, 15)
	if !slices.Equal(got, []int{0, 1, 1, 2, 3, 5, 8, 3, 1, 4, 5, 9, 4, 3, 7}) {
		t.Errorf("%v", got)
	}
	// the Pisano period of 10 is 60
	f.Reset()
	a := take(f, 60)
	b := take(f, 60)
	if !slices.Equal(a, b) {
		t.Error("not periodic with period 60")
	}
	if got := take(NewFibonacci(1000), 17); got[16] != 987 {
		t.Errorf("F16 = %d", got[16])
	}
}

func TestCollatz(t *testing.T) {
	c := NewCollatz(6)
	var got []int
	for !c.Done() {
		got = append(got, c.Generate())
	}
	if !slices.Equal(got, []int{6, 3, 10, 5, 16, 8, 4, 2, 1}) {
		t.Errorf("%v", got)
	}
	l := NewCollatzLoop(1)
	if !l.Continuous() || l.Generate() != 1 || l.Generate() != 1 {
		t.Error("loop of 1")
	}
}

func TestThueMorseAndPrimes(t *testing.T) {
	if got := take(NewThueMorse(), 16); !slices.Equal(got, []int{0, 1, 1, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0}) {
		t.Errorf("%v", got)
	}
	p := NewPrimes()
	if got := take(p, 10); !slices.Equal(got, []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}) {
		t.Errorf("%v", got)
	}
	p.Reset()
	if p.Generate() != 2 {
		t.Error("reset")
	}
}

func TestInvalid(t *testing.T) {
	for name, f := range map[string]func(){
		"mod 1":     func() { NewFibonacci(1) },
		"collatz 0": func() { NewCollatz(0) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s should panic", name)
				}
			}()
			f()
		}()
	}
}
