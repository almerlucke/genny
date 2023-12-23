package main

import (
	"genny"
	"genny/and"
	"genny/bucket"
	"genny/flatten"
	"genny/function"
	"genny/or"
	"genny/repeat"
	"genny/sequence"
	"genny/transform"
	"genny/walk"
	"log"
	"math/rand"
)

func main() {
	var g genny.Generator[float64] = sequence.New(1.0, 2.0, 3.0)
	for !g.Done() {
		log.Printf("sequence: %f", g.NextValue())
	}

	g = function.New(nil, func(_ any) float64 { return rand.Float64() })
	for i := 0; i < 10; i++ {
		log.Printf("function: %f", g.NextValue())
	}

	g = repeat.New[float64](sequence.New(2.0, 3.0, 4.0), 5, 8)
	for !g.Done() {
		log.Printf("repeat: %f", g.NextValue())
	}

	g = and.New[float64](
		repeat.New[float64](sequence.New(2.0, 3.0, 4.0), 4, 4),
		repeat.New[float64](sequence.New(8.0, 9.0, 10.0), 4, 4),
	)
	for !g.Done() {
		log.Printf("and: %f", g.NextValue())
	}

	g = bucket.New(bucket.Random, 2.0, 3.0, 4.0, 5.0)
	for i := 0; i < 10; i++ {
		log.Printf("bucket random: %f", g.NextValue())
	}

	g = bucket.NewContinuous(bucket.Indexed, 2.0, 3.0, 4.0, 5.0)
	for i := 0; i < 10; i++ {
		log.Printf("bucket indexed: %f", g.NextValue())
	}

	g = or.New[float64](or.Indexed, sequence.New(2.0, 3.0, 4.0), sequence.New(12.0, 13.0, 14.0))
	for !g.Done() {
		log.Printf("or indexed: %f", g.NextValue())
	}

	g = or.NewContinuous[float64](or.Indexed, sequence.NewContinuous(2.0, 3.0, 4.0), sequence.NewContinuous(12.0, 13.0, 14.0))
	for i := 0; i < 20; i++ {
		log.Printf("or indexed continuous: %f", g.NextValue())
	}

	g = flatten.NewFlatten[float64](sequence.New[[]float64]([]float64{1.0, 2.0, 3.0}, []float64{4.0, 5.0, 6.0}))
	for !g.Done() {
		log.Printf("flatten: %f", g.NextValue())
	}

	g = walk.New[float64](walk.NewMatrix(
		[]int{3, 3},
		[]float64{
			1.0, 2.0, 3.0,
			4.0, 5.0, 6.0,
			7.0, 8.0, 9.0,
		}))
	for i := 0; i < 10; i++ {
		log.Printf("walk: %f", g.NextValue())
	}

	var f transform.Function[float64] = func(v float64) float64 { return v + 12 }
	g = transform.New[float64](sequence.New(2.0, 3.0, 4.0), f)
	for !g.Done() {
		log.Printf("transform: %f", g.NextValue())
	}
}
