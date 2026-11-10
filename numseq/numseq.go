// Package numseq has number sequences that give structure: Fibonacci, Collatz, Thue-Morse and the primes.
// Map them to pitches, durations or gates (for instance with a modulo or a scale degree)
package numseq

import (
	"math/bits"

	"github.com/almerlucke/genny/sequence"
)

// Fibonacci generates 0 1 1 2 3 5 8 ... reduced modulo mod, which makes it a repeating pattern
type Fibonacci struct {
	mod  int
	a, b int
}

// NewFibonacci creates a Fibonacci generator, values are in [0, mod). It panics if mod is less than 2
func NewFibonacci(mod int) *Fibonacci {
	if mod < 2 {
		panic("numseq: mod must be at least 2")
	}
	return &Fibonacci{mod: mod, b: 1}
}

func (f *Fibonacci) Generate() int {
	v := f.a
	f.a, f.b = f.b, (f.a+f.b)%f.mod
	return v
}

func (f *Fibonacci) Continuous() bool { return true }

func (f *Fibonacci) Done() bool { return false }

func (f *Fibonacci) Reset() { f.a, f.b = 0, 1 }

func collatz(start int) []int {
	if start < 1 {
		panic("numseq: collatz start must be at least 1")
	}
	vals := []int{start}
	for n := start; n != 1; {
		if n%2 == 0 {
			n /= 2
		} else {
			n = 3*n + 1
		}
		vals = append(vals, n)
	}
	return vals
}

// NewCollatz plays the Collatz trajectory of start (halve if even, else 3n+1) until it reaches 1, then it
// is done. It panics if start is less than 1
func NewCollatz(start int) *sequence.Sequence[int] { return sequence.New(collatz(start)...) }

// NewCollatzLoop is like NewCollatz but loops
func NewCollatzLoop(start int) *sequence.Sequence[int] { return sequence.NewLoop(collatz(start)...) }

// ThueMorse generates 0 1 1 0 1 0 0 1 ...: a sequence that never settles into a repeating pattern
// but has no long runs, a good source of gates
type ThueMorse struct {
	n uint
}

// NewThueMorse creates a Thue-Morse generator
func NewThueMorse() *ThueMorse { return &ThueMorse{} }

func (t *ThueMorse) Generate() int {
	v := bits.OnesCount(t.n) % 2
	t.n++
	return v
}

func (t *ThueMorse) Continuous() bool { return true }

func (t *ThueMorse) Done() bool { return false }

func (t *ThueMorse) Reset() { t.n = 0 }

// Primes generates 2 3 5 7 11 ...
type Primes struct {
	last int
}

// NewPrimes creates a generator of the prime numbers
func NewPrimes() *Primes { return &Primes{} }

func (p *Primes) Generate() int {
	n := max(p.last+1, 2)
	for !isPrime(n) {
		n++
	}
	p.last = n
	return n
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for d := 2; d*d <= n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}

func (p *Primes) Continuous() bool { return true }

func (p *Primes) Done() bool { return false }

func (p *Primes) Reset() { p.last = 0 }
