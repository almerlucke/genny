package groove

import (
	"testing"

	"github.com/almerlucke/genny/constant"
)

func TestSwing(t *testing.T) {
	s := NewSwing(0.25)
	for i := range 6 {
		want := 0.0
		if i%2 == 1 {
			want = 0.25
		}
		if got := s.Generate(); got != want {
			t.Fatalf("step %d: %v, want %v", i, got, want)
		}
	}
	s.Reset()
	if s.Generate() != 0 {
		t.Error("reset")
	}
}

func TestHumanizeBounds(t *testing.T) {
	h := NewHumanize(0.1)
	s := NewSwing(0.2).Humanize(0.05)
	j := Jitter(constant.New(0.5), 0.1)
	for i := range 1000 {
		if v := h.Generate(); v < -0.1 || v > 0.1 {
			t.Fatal(v)
		}
		base := 0.0
		if i%2 == 1 {
			base = 0.2
		}
		if v := s.Generate(); v < base-0.05 || v > base+0.05 {
			t.Fatal(v)
		}
		if v := j.Generate(); v < 0.4 || v > 0.6 {
			t.Fatal(v)
		}
	}
}
