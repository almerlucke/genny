package main

import (
	"genny"
	"genny/function"
	"genny/repeat"
	"genny/sequence"
	"log"
	"math/rand"
)

func main() {
	var g genny.Generator[float64] = sequence.NewSequence(1.0, 2.0, 3.0)
	for !g.Done() {
		log.Printf("sequence: %f", g.NextValue())
	}

	g = function.NewFunction(nil, func(_ any) float64 { return rand.Float64() })
	for i := 0; i < 10; i++ {
		log.Printf("function: %f", g.NextValue())
	}

	g = repeat.NewRepeat[float64](sequence.NewSequence(2.0, 3.0, 4.0), 5, 8)
	for !g.Done() {
		log.Printf("repeat: %f", g.NextValue())
	}
}
