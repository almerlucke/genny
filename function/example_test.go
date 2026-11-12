package function_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/function"
)

func ExampleNew() {
	n := 0
	g := function.New(func() int { n += 2; return n })
	fmt.Println(g.Generate(), g.Generate(), g.Generate())
	// Output: 2 4 6
}

func ExampleNewWithContext() {
	g := function.NewWithContext("x", func(ctx any) string { return ctx.(string) + "!" })
	fmt.Println(g.Generate())
	// Output: x!
}

func ExampleNewRandom() {
	genny.Seed(1)
	g := function.NewRandom(10, 20)
	ok := true
	for i := 0; i < 100; i++ {
		if v := g.Generate(); v < 10 || v >= 20 {
			ok = false
		}
	}
	fmt.Println(ok)
	// Output: true
}
