package main

import (
	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/conv"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/template"
	"log"
)

func main() {
	var g genny.Generator[float64] = sequence.New(1.0, 2.0, 3.0)

	t := template.Template{
		"test": conv.ToAny(g),
	}

	for !t.Done() {
		maps := t.Generate()
		log.Printf("maps: %v", maps)
	}
}
