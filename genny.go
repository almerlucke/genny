package genny

import (
	"math/rand/v2"
	"sync"
)

type Generator[T any] interface {
	// Generate generates a value
	Generate() T
	// Continuous should return false if Done() and Reset() are used by this generator,
	// this allows for generators that can have an end state
	Continuous() bool
	// Done returns true if the generator is done generating values
	Done() bool
	// Reset the generator
	Reset()
}

// The generators draw their random numbers from one stream. Seed makes it repeatable: after the
// same seed the same calls give the same values. Without a seed every run is different. The stream
// is locked, so generators may be used from several goroutines, but then the order of the draws
// (and with it the values) depends on who comes first: seed and draw from one goroutine to repeat
var (
	rngMu sync.Mutex
	rng   = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
)

// Seed seeds the random stream of the generators
func Seed(seed uint64) {
	rngMu.Lock()
	rng = rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	rngMu.Unlock()
}

// Float64 returns a random number in [0, 1) from the stream of the generators
func Float64() float64 {
	rngMu.Lock()
	defer rngMu.Unlock()
	return rng.Float64()
}

// Intn returns a random number in [0, n) from the stream of the generators
func Intn(n int) int {
	rngMu.Lock()
	defer rngMu.Unlock()
	return rng.IntN(n)
}

// Shuffle shuffles n elements with the stream of the generators
func Shuffle(n int, swap func(i, j int)) {
	rngMu.Lock()
	defer rngMu.Unlock()
	rng.Shuffle(n, swap)
}
