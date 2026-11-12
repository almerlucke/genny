package smoothstep_test

import (
	"fmt"

	"github.com/almerlucke/genny/float/shape/shapers/smoothstep"
)

func ExampleNew() {
	for _, k := range []smoothstep.Kind{smoothstep.Smoothstep, smoothstep.Smootherstep, smoothstep.Inverse} {
		s := smoothstep.New(k, 0, 1)
		fmt.Printf("%.4f %.4f\n", s.Shape(0.25), s.Shape(2))
	}
	// Output:
	// 0.1562 1.0000
	// 0.1035 1.0000
	// 0.3264 1.0000
}
