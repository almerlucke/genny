package bucket_test

import (
	"fmt"
	"sort"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/bucket"
)

// By default a bucket hands out every value once, in a random order.
func ExampleNew() {
	genny.Seed(1)
	b := bucket.New(60, 62, 64, 65, 67)
	var out []int
	for !b.Done() {
		out = append(out, b.Generate())
	}
	fmt.Println(len(out))
	sort.Ints(out)
	fmt.Println(out)
	// Output:
	// 5
	// [60 62 64 65 67]
}

// In Random mode values are drawn with replacement and the bucket never ends.
func ExampleBucket_SetMode() {
	genny.Seed(1)
	b := bucket.New("a", "b", "c").SetMode(bucket.Random)
	fmt.Println(b.Continuous(), b.Done())
	// Output: true false
}

func ExampleNewLoop() {
	genny.Seed(2)
	b := bucket.NewLoop(1, 2, 3)
	sum := 0
	for i := 0; i < 6; i++ { // two full shuffled rounds
		sum += b.Generate()
	}
	fmt.Println(sum, b.Done())
	// Output: 12 false
}
