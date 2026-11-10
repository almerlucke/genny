package transform_test

import (
	"testing"

	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/transform"
)

type counter struct{ n, resets int }

func (c *counter) Transform(v int) int { c.n++; return v*10 + c.n }
func (c *counter) Reset()              { c.n = 0; c.resets++ }

func TestTransformFunction(t *testing.T) {
	tr := transform.New[int](sequence.New(1, 2), func(v int) int { return v * v })
	if tr.Continuous() || tr.Done() {
		t.Fatal("state follows wrapped generator")
	}
	if tr.Generate() != 1 || tr.Generate() != 4 {
		t.Fatal("transformed values")
	}
	if !tr.Done() {
		t.Fatal("done follows wrapped generator")
	}
	tr.Reset()
	if tr.Done() || tr.Generate() != 1 {
		t.Fatal("reset")
	}
}

func TestTransformerReset(t *testing.T) {
	c := &counter{}
	tr := transform.NewWithTransformer[int](sequence.NewLoop(1, 2), c)
	if !tr.Continuous() {
		t.Fatal("continuous follows wrapped generator")
	}
	if tr.Generate() != 11 || tr.Generate() != 22 {
		t.Fatal("stateful transformer")
	}
	tr.Reset()
	if c.resets != 1 || tr.Generate() != 11 {
		t.Fatal("transformer must be reset too")
	}
}
