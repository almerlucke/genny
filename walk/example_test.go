package walk_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/walk"
)

// Walk takes random single steps through a matrix, so consecutive values are
// always neighbours in one dimension.
func ExampleNew() {
	genny.Seed(3)
	m := walk.NewMatrix([]int{5}, []int{10, 20, 30, 40, 50})
	w := walk.New[int](m)
	prev := w.Generate()
	ok := true
	for i := 0; i < 50; i++ {
		v := w.Generate()
		if d := v - prev; d != 10 && d != -10 {
			ok = false
		}
		prev = v
	}
	fmt.Println(ok)
	// Output: true
}
