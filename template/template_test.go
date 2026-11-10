package template_test

import (
	"reflect"
	"testing"

	"github.com/almerlucke/genny/constant"
	"github.com/almerlucke/genny/sequence"
	"github.com/almerlucke/genny/template"
)

type m = map[string]any

func TestPlainTemplateIsCopied(t *testing.T) {
	tpl := template.Template{"a": 1, "b": "x", "n": template.Template{"c": 2.5}}
	got := tpl.Generate()
	want := []map[string]any{{"a": 1, "b": "x", "n": m{"c": 2.5}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestGeneratorsAreReplacedByTheirValues(t *testing.T) {
	tpl := template.Template{
		"n":   sequence.New(1, 2, 3),
		"c":   constant.New("k"),
		"sub": template.Template{"d": sequence.NewLoop(10, 20)},
	}
	if tpl.Continuous() || tpl.Done() {
		t.Fatal("a sequence in the template makes it non continuous, fresh is not done")
	}
	for i, want := range []int{1, 2, 3} {
		got := tpl.Generate()
		sub := []int{10, 20, 10}[i]
		if !reflect.DeepEqual(got, []map[string]any{{"n": want, "c": "k", "sub": m{"d": sub}}}) {
			t.Fatalf("round %d got %v", i, got)
		}
	}
	if !tpl.Done() {
		t.Fatal("done as soon as one generator is done")
	}
	tpl.Reset()
	if tpl.Done() {
		t.Fatal("reset clears done")
	}
	if got := tpl.Generate(); got[0]["n"] != 1 || got[0]["sub"].(map[string]any)["d"] != 10 {
		t.Fatalf("reset must reach generators and sub templates: %v", got)
	}
}

func TestContinuousTemplate(t *testing.T) {
	tpl := template.Template{"a": constant.New(1), "b": 2, "s": template.Template{"c": sequence.NewLoop(1)}}
	if !tpl.Continuous() || tpl.Done() {
		t.Fatal("only continuous generators")
	}
	if !(template.Template{}).Continuous() || (template.Template{}).Done() {
		t.Fatal("empty template")
	}
}

func TestSlicesSplitTheOutput(t *testing.T) {
	tpl := template.Template{
		"a": constant.New([]int{1, 2, 3}),
		"b": constant.New([]string{"x", "y"}),
		"c": "plain",
		"s": template.Template{"d": constant.New([4]int{9, 8, 7, 6})},
	}
	got := tpl.Generate()
	want := []map[string]any{
		{"a": 1, "b": "x", "c": "plain", "s": m{"d": 9}},
		{"a": 2, "b": "y", "c": "plain", "s": m{"d": 8}},
		{"a": 3, "b": "y", "c": "plain", "s": m{"d": 7}}, // shorter slices stick to their last value
		{"a": 3, "b": "y", "c": "plain", "s": m{"d": 6}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

func TestEmptySliceDoesNotPanic(t *testing.T) {
	tpl := template.Template{"a": constant.New([]int{}), "b": constant.New([]int{1, 2})}
	got := tpl.Generate()
	if len(got) != 2 || len(got[0]["a"].([]int)) != 0 || got[1]["b"] != 2 {
		t.Fatalf("got %v", got)
	}
	if got := (template.Template{"a": constant.New([]int{})}).Generate(); len(got) != 1 {
		t.Fatalf("an empty slice still gives one map, got %v", got)
	}
}

func TestNilValueDoesNotPanic(t *testing.T) {
	tpl := template.Template{"a": nil, "b": 1}
	got := tpl.Generate()
	if !reflect.DeepEqual(got, []map[string]any{{"a": nil, "b": 1}}) {
		t.Fatalf("got %v", got)
	}
	if !tpl.Continuous() || tpl.Done() {
		t.Fatal("nil is a plain value")
	}
	tpl.Reset()
}

func TestParameters(t *testing.T) {
	tpl := template.Template{
		"freq": template.NewParameter("freq", 440.0),
		"sub":  template.Template{"f2": template.NewParameter("freq", 440.0), "amp": template.NewParameter("amp", 1)},
	}
	tpl.SetParameter("freq", 220.0)
	got := tpl.Generate()[0]
	if got["freq"] != 220.0 || got["sub"].(map[string]any)["f2"] != 220.0 || got["sub"].(map[string]any)["amp"] != 1 {
		t.Fatalf("got %v", got)
	}

	tpl.SetParameters([]*template.Parameter{template.NewParameter("amp", 0.5), template.NewParameter("freq", 110.0)})
	got = tpl.Generate()[0]
	if got["freq"] != 110.0 || got["sub"].(map[string]any)["amp"] != 0.5 {
		t.Fatalf("got %v", got)
	}

	// a parameter value can be a slice, which splits
	tpl.SetParameter("freq", []float64{1, 2})
	if out := tpl.Generate(); len(out) != 2 || out[1]["freq"] != 2.0 {
		t.Fatalf("got %v", out)
	}

	// a parameter that is not a pointer is stored back in the template
	tv := template.Template{"p": template.Parameter{Name: "p", Value: 1}}
	tv.SetParameter("p", 2)
	if got := tv.Generate()[0]["p"]; got != 2 {
		t.Fatalf("non pointer parameter not updated: %v", got)
	}
	// an unknown name changes nothing
	tv.SetParameter("other", 3)
	if got := tv.Generate()[0]["p"]; got != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestEmptyTemplate(t *testing.T) {
	got := template.Template{}.Generate()
	if len(got) != 1 || len(got[0]) != 0 {
		t.Fatalf("got %v", got)
	}
}
