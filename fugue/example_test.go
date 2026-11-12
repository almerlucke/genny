package fugue_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/fugue"
	"github.com/almerlucke/genny/scale"
)

// A tonal answer: the subject starts on the tonic and leaps to the dominant (D to A), the answer starts on the
// dominant and leaps up a fourth to the tonic (A to D), where a real answer would go to E
func ExampleAnswer() {
	subject := fugue.NewSubject([]int{0, 4, 5, 3}, []float64{4, 2, 1, 1})
	key := scale.New(62, scale.Minor) // D minor
	fmt.Println("tonal:", fugue.Answer(subject, key, fugue.Tonal).Degrees)
	fmt.Println("real: ", fugue.Answer(subject, key, fugue.Real).Degrees)
	// Output:
	// tonal: [4 7 9 7]
	// real:  [4 8 9 7]
}

// An exposition in three voices, then a stretto and a close. Check verifies that every entry is the
// subject or the answer, moved the right way
func ExampleExposition() {
	genny.Seed(3)
	key := scale.New(62, scale.Minor)
	subject := fugue.NewSubject([]int{0, 4, 5, 3, 1, 4, 3, 2, 1, 0}, []float64{8, 4, 2, 2, 4, 2, 2, 2, 2, 4})
	f, err := fugue.Exposition(subject, 3, key, fugue.Config{})
	if err != nil {
		fmt.Println(err)
	}
	if err := f.Stretto(16, nil); err != nil {
		fmt.Println(err)
	}
	f.Close(fugue.CloseConfig{LeadingTone: true})
	for _, e := range f.Entries {
		fmt.Printf("bar %d: voice %d, %s\n", f.Bar(e.Start), e.Voice, e.Kind)
	}
	fmt.Println("problems:", len(f.Check()))
	// Output:
	// bar 1: voice 0, subject
	// bar 3: voice 1, answer
	// bar 5: voice 2, subject
	// bar 7: voice 0, subject
	// bar 8: voice 1, answer
	// bar 9: voice 2, subject
	// problems: 0
}
