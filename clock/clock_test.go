package clock

import "testing"

func TestDivider(t *testing.T) {
	d := NewDivider(3, 0)
	var got []bool
	for range 7 {
		got = append(got, d.Generate())
	}
	want := []bool{true, false, false, true, false, false, true}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%v", got)
		}
	}
	d = NewDivider(3, 1) // offset moves the ticks earlier
	if d.Generate() || d.Generate() || !d.Generate() {
		t.Error("offset")
	}
}

func TestPoly(t *testing.T) {
	p := NewPoly(3, 4)
	if p.Cycle() != 12 {
		t.Fatal(p.Cycle())
	}
	both := 0
	for i := range 24 {
		o := p.Generate()
		if o[0] != (i%3 == 0) || o[1] != (i%4 == 0) {
			t.Fatalf("step %d: %v", i, o)
		}
		if o[0] && o[1] {
			both++
		}
	}
	if both != 2 {
		t.Errorf("realigned %d times in 24 steps", both)
	}
}

func TestInvalid(t *testing.T) {
	for name, f := range map[string]func(){
		"divider 0":     func() { NewDivider(0, 0) },
		"poly length 0": func() { NewPoly(3, 0) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s should panic", name)
				}
			}()
			f()
		}()
	}
}
