package main

import (
	"github.com/almerlucke/genny/float"
	"github.com/almerlucke/genny/float/export/plot"
	"github.com/almerlucke/genny/float/phasor"
	"github.com/almerlucke/genny/float/shape"
	"github.com/almerlucke/genny/float/shape/shapers/lookup"
	"log"
)

func main() {
	ph := phasor.New(1000, 44100.0, 0.0)
	sh := shape.New(float.ToFrame(ph), lookup.NewSineTable(256))

	err := plot.Frames(
		sh,
		100,
		plot.Centimeter*10, plot.Centimeter*5,
		"sine.png",
	)
	if err != nil {
		log.Fatalf("plot error: %v", err)
	}
}
