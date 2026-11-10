package constant

import "testing"

func TestConstant(t *testing.T) {
	c := New("x")
	for i := 0; i < 3; i++ {
		if c.Generate() != "x" {
			t.Fatal("constant changed")
		}
	}
	if !c.Continuous() || c.Done() {
		t.Fatal("constant is continuous and never done")
	}
	c.Reset()
	if c.Generate() != "x" {
		t.Fatal("Reset must not change the value")
	}
	c.SetValue("y")
	if c.Generate() != "y" {
		t.Fatal("SetValue")
	}
}
