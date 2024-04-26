package main

import (
	"log"
	"math/rand"
)

func main() {
	n1 := 4
	n2 := 5
	dif := n2 + 1 - n1
	f := func() int {
		return rand.Intn(dif) + n1
	}

	for _ = range 10 {
		log.Printf("rand: %d", f())
	}
}
