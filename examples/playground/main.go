package main

import "log"

func main() {
	shapers := []float64{0.0, 1.0, 2.0, 3.0}
	index := 0.001
	fx := index * float64(len(shapers)-1)
	ix1 := int(fx)
	ix2 := ix1 + 1

	if ix2 >= len(shapers) {
		ix2 = ix1
	}

	sx1 := shapers[ix1]
	sx2 := shapers[ix2]

	result := sx1 + (fx-float64(ix1))*(sx2-sx1)

	log.Printf("result: %v", result)
}
