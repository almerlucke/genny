package constant_test

import (
	"fmt"

	"github.com/almerlucke/genny/constant"
)

func ExampleNew() {
	g := constant.New(440.0)
	fmt.Println(g.Generate(), g.Generate(), g.Continuous(), g.Done())
	// Output: 440 440 true false
}
