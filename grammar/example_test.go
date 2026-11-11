package grammar_test

import (
	"fmt"

	"github.com/almerlucke/genny"
	"github.com/almerlucke/genny/grammar"
)

// A melodic phrase in A minor (scale degrees): a phrase is two or three motifs, a motif is a
// rising or falling figure that may repeat itself
func ExampleGrammar_melody() {
	genny.Seed(11)
	g := grammar.New[int]().
		Rule("phrase",
			grammar.Choice(2, grammar.Ref[int]("motif"), grammar.Ref[int]("motif")),
			grammar.Choice(1, grammar.Ref[int]("motif"), grammar.Ref[int]("answer"), grammar.Ref[int]("motif"))).
		Rule("motif",
			grammar.Seq(0, 2, 4),
			grammar.Seq(4, 2, 0),
			grammar.Choice(1, grammar.Term(2), grammar.Term(4), grammar.Ref[int]("motif"))).
		Rule("answer", grammar.Seq(7, 4, 2, 0))
	for range 3 {
		fmt.Println(g.Expand("phrase", 4))
	}
	// Output:
	// [2 4 4 2 0 4 2 0]
	// [2 4 4 2 0 2 4 2 4 2 4 0 2 4]
	// [4 2 0 0 2 4]
}

// A bar of rhythm in sixteenths: a bar is four beats, a beat is a quarter note, two eighths or
// four sixteenths, or a dotted eighth and a sixteenth (the numbers are lengths in sixteenths, so a
// bar always adds up to 16)
func ExampleGrammar_rhythm() {
	genny.Seed(3)
	g := grammar.New[int]().
		Rule("bar", grammar.Choice(1, grammar.Ref[int]("beat"), grammar.Ref[int]("beat"), grammar.Ref[int]("beat"), grammar.Ref[int]("beat"))).
		Rule("beat",
			grammar.Choice(2, grammar.Term(4)),
			grammar.Choice(2, grammar.Term(2), grammar.Term(2)),
			grammar.Choice(1, grammar.Term(1), grammar.Term(1), grammar.Term(1), grammar.Term(1)),
			grammar.Choice(1, grammar.Term(3), grammar.Term(1)))
	bar := g.Expand("bar", 2)
	sum := 0
	for _, d := range bar {
		sum += d
	}
	fmt.Println(bar, "adds up to", sum)
	// Output:
	// [3 1 2 2 4 2 2] adds up to 16
}

// A player makes a new derivation every cycle: a call and a response, each in two versions
func ExamplePlayer() {
	genny.Seed(2)
	g := grammar.New[string]().
		Rule("call", grammar.Seq("do", "re"), grammar.Seq("do", "mi")).
		Rule("response", grammar.Seq("sol"), grammar.Seq("la")).
		Rule("tune", grammar.Choice(1, grammar.Ref[string]("call"), grammar.Ref[string]("response")))
	p := grammar.NewPlayer(g, "tune", 3)
	for range 4 {
		fmt.Println(p.Generate(), p.Generate(), p.Generate())
	}
	// Output:
	// do mi la
	// do mi sol
	// do re la
	// do re sol
}
