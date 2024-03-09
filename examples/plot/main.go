package main

import (
	"github.com/almerlucke/genny/float/conversions"
	"github.com/almerlucke/genny/float/phasor"
	"github.com/almerlucke/genny/float/plot"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/lookup"
	"gonum.org/v1/plot/vg"
	"log"
)

func main() {
	ph := phasor.New(1000, 44100.0, 0.0)
	sh := shape.New(conversions.ToVec(ph), 1, lookup.NewSineTable(256))

	err := plot.Plot(
		sh,
		1,
		100,
		vg.Centimeter*10, vg.Centimeter*5,
		"sine.png",
	)
	if err != nil {
		log.Fatalf("plot error: %v", err)
	}
}
