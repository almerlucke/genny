package conv_test

import (
	"fmt"
	"testing"

	"github.com/almerlucke/genny/conv"
	"github.com/almerlucke/genny/sequence"
)

type itoa struct{}

func (itoa) Convert(i int) string { return fmt.Sprintf("<%d>", i) }

func TestConvert(t *testing.T) {
	c := conv.New[int, string](sequence.New(1, 2), itoa{})
	if c.Continuous() || c.Done() {
		t.Fatal("state comes from the wrapped generator")
	}
	if c.Generate() != "<1>" || c.Generate() != "<2>" {
		t.Fatal("conversion")
	}
	if !c.Done() {
		t.Fatal("done follows wrapped generator")
	}
	c.Reset()
	if c.Done() || c.Generate() != "<1>" {
		t.Fatal("reset follows wrapped generator")
	}
	if !conv.New[int, string](sequence.NewLoop(1), itoa{}).Continuous() {
		t.Fatal("continuous follows wrapped generator")
	}
}

func TestToAny(t *testing.T) {
	a := conv.ToAny[int](sequence.New(7))
	v := a.Generate()
	if i, ok := v.(int); !ok || i != 7 {
		t.Fatalf("got %#v", v)
	}
}
