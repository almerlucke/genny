package main

import (
	"genny"
	"genny/sequence"
	"log"
)

func main() {
	var g genny.Generator[float64] = sequence.NewSequence(2.3, 2.2, 3.4)

	for !g.Done() {
		log.Printf("next value: %f", g.NextValue())
	}
}
