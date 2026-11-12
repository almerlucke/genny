package polyrhythm

import (
	"fmt"
	"math/bits"
)

// Frac is an exact rational number, used for every time and tempo in this package so that nothing
// drifts. Times are measured in beats, the unit of the tempo in beats per minute (a quarter note in
// 4/4): 1/2 is an eighth note, 2/3 a quarter note triplet, 4/5 a quintuplet over a bar of four beats.
// A step of a note.Event is 1/StepsPerBeat of a beat, so a time t becomes the step t*StepsPerBeat.
//
// A Frac is normalized by every operation (lowest terms, positive denominator), so results can be
// compared with == (the zero value Frac{} is 0 too, but is only == to Int(0) after an operation: use Cmp
// or IsZero for values you made by hand). Arithmetic panics on a zero denominator and on int64
// overflow, which cannot happen with musical numbers
type Frac struct {
	Num, Den int64 // Den == 0 is read as 1
}

// F returns n/d in lowest terms. It panics if d is 0
func F(n, d int64) Frac {
	if d == 0 {
		panic("polyrhythm: zero denominator")
	}
	if d < 0 {
		n, d = -n, -d
	}
	g := gcd(abs(n), d)
	return Frac{n / g, d / g}
}

// Int returns n as a Frac
func Int(n int64) Frac { return Frac{n, 1} }

func abs(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	if a == 0 {
		return 1
	}
	return a
}

func mul(a, b int64) int64 {
	neg := (a < 0) != (b < 0)
	hi, lo := bits.Mul64(uint64(abs(a)), uint64(abs(b)))
	if hi != 0 || lo > 1<<63-1 {
		panic("polyrhythm: overflow")
	}
	if neg {
		return -int64(lo)
	}
	return int64(lo)
}

func add(a, b int64) int64 {
	c := a + b
	if (a > 0 && b > 0 && c < 0) || (a < 0 && b < 0 && c >= 0) {
		panic("polyrhythm: overflow")
	}
	return c
}

func (a Frac) norm() Frac {
	if a.Den == 0 {
		return Frac{a.Num, 1}
	}
	return F(a.Num, a.Den)
}

// Add returns a+b
func (a Frac) Add(b Frac) Frac {
	a, b = a.norm(), b.norm()
	g := gcd(a.Den, b.Den)
	return F(add(mul(a.Num, b.Den/g), mul(b.Num, a.Den/g)), mul(a.Den, b.Den/g))
}

// Sub returns a-b
func (a Frac) Sub(b Frac) Frac { return a.Add(b.Neg()) }

// Neg returns -a
func (a Frac) Neg() Frac { a = a.norm(); return Frac{-a.Num, a.Den} }

// Mul returns a*b
func (a Frac) Mul(b Frac) Frac {
	a, b = a.norm(), b.norm()
	g1, g2 := gcd(abs(a.Num), b.Den), gcd(abs(b.Num), a.Den)
	return F(mul(a.Num/g1, b.Num/g2), mul(a.Den/g2, b.Den/g1))
}

// Div returns a/b. It panics if b is 0
func (a Frac) Div(b Frac) Frac {
	b = b.norm()
	if b.Num == 0 {
		panic("polyrhythm: division by zero")
	}
	return a.Mul(F(b.Den, b.Num))
}

// MulInt returns a*n
func (a Frac) MulInt(n int64) Frac { return a.Mul(Int(n)) }

// Cmp compares a and b: -1, 0 or +1
func (a Frac) Cmp(b Frac) int {
	d := a.Sub(b).Num
	switch {
	case d < 0:
		return -1
	case d > 0:
		return 1
	}
	return 0
}

// IsZero tells if a is 0
func (a Frac) IsZero() bool { return a.Num == 0 }

// Sign is -1, 0 or +1
func (a Frac) Sign() int { return a.Cmp(Frac{}) }

// IsInt tells if a is a whole number
func (a Frac) IsInt() bool { return a.norm().Den == 1 }

// Floor is the largest integer not above a
func (a Frac) Floor() int64 {
	a = a.norm()
	q := a.Num / a.Den
	if a.Num%a.Den != 0 && a.Num < 0 {
		q--
	}
	return q
}

// Ceil is the smallest integer not below a
func (a Frac) Ceil() int64 { return -a.Neg().Floor() }

// Float converts to a float64 (one rounding, done once at the end of exact arithmetic)
func (a Frac) Float() float64 { a = a.norm(); return float64(a.Num) / float64(a.Den) }

// String is "n/d", or "n" for a whole number
func (a Frac) String() string {
	a = a.norm()
	if a.Den == 1 {
		return fmt.Sprint(a.Num)
	}
	return fmt.Sprintf("%d/%d", a.Num, a.Den)
}

func lcm(a, b int64) int64 { return mul(a/gcd(a, b), b) }

// GCD is the greatest common divisor of two positive fractions: the largest f such that both are
// whole multiples of f. gcd(a/b, c/d) = gcd(a, c) / lcm(b, d)
func GCD(a, b Frac) Frac {
	a, b = a.norm(), b.norm()
	return F(gcd(abs(a.Num), abs(b.Num)), lcm(a.Den, b.Den))
}

// LCM is the least common multiple of two positive fractions: the smallest time at which both
// repeat. lcm(a/b, c/d) = lcm(a, c) / gcd(b, d)
func LCM(a, b Frac) Frac {
	a, b = a.norm(), b.norm()
	return F(lcm(abs(a.Num), abs(b.Num)), gcd(a.Den, b.Den))
}
