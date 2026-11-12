package template_test

import (
	"fmt"

	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/template"
)

// A Template maps names to generators, constants and parameters. Generate
// produces one map per step; slices are expanded and shorter ones repeat their
// last element.
func ExampleTemplate() {
	t := template.Template{
		"pitch": sequence.NewLoop(60, 64, 67),
		"vel":   template.NewParameter("vel", []int{90, 70}),
		"inst":  "piano",
	}
	for i := 0; i < 2; i++ {
		fmt.Println(t.Generate())
	}
	// Output:
	// [map[inst:piano pitch:60 vel:90] map[inst:piano pitch:60 vel:70]]
	// [map[inst:piano pitch:64 vel:90] map[inst:piano pitch:64 vel:70]]
}

func ExampleTemplate_SetParameter() {
	t := template.Template{"vel": template.NewParameter("vel", 64)}
	t.SetParameter("vel", 100)
	fmt.Println(t.Generate())
	// Output: [map[vel:100]]
}
