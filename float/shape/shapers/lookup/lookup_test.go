package lookup

import (
	"math"
	"reflect"
	"testing"
)

func eq(a, b Table) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(a[i]-b[i]) > 1e-12 {
			return false
		}
	}
	return true
}

func TestTables(t *testing.T) {
	tests := []struct {
		name  string
		table Table
		want  Table
	}{
		{"sine", NewSineTable(4), Table{0, 1, 0, -1, 0}},
		{"uni sine", NewUniSineTable(4), Table{0.5, 1, 0.5, 0, 0.5}},
		{"cosine", NewCosineTable(4), Table{1, 0, -1, 0, 1}},
		{"uni cosine", NewUniCosineTable(4), Table{1, 0.5, 0, 0.5, 1}},
		{"saw", NewSawtoothTable(4), Table{-1, -0.5, 0, 0.5, -1}},
		{"uni saw", NewUniSawtoothTable(4), Table{0, 0.25, 0.5, 0.75, 0}},
		{"triangle", NewTriangleTable(4), Table{0, 1, 0, -1, 0}},
		{"uni triangle", NewUniTriangleTable(4), Table{0.5, 1, 0.5, 0, 0.5}},
		{"square", NewSquareTable(4), Table{1, 1, -1, -1, 1}},
		{"uni square", NewUniSquareTable(4), Table{1, 1, 0, 0, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !eq(tt.table, tt.want) {
				t.Fatalf("got %v want %v", tt.table, tt.want)
			}
		})
	}
}

func TestFillModes(t *testing.T) {
	f := func(i, n int) float64 { return float64(i * 10) }
	if got := NewWithFunc(3, First, f); !reflect.DeepEqual(got, Table{0, 10, 20, 0}) {
		t.Errorf("First: %v", got)
	}
	if got := NewWithFunc(3, Last, f); !reflect.DeepEqual(got, Table{0, 10, 20, 20}) {
		t.Errorf("Last: %v", got)
	}
	if got := NewWithFunc(3, Custom, f, 99); !reflect.DeepEqual(got, Table{0, 10, 20, 99}) {
		t.Errorf("Custom: %v", got)
	}
	if got := NewWithFunc(3, Custom, f); !reflect.DeepEqual(got, Table{0, 10, 20, 0}) {
		t.Errorf("Custom without value: %v", got)
	}
	// f gets the index and the size
	NewWithFunc(5, First, func(i, n int) float64 {
		if n != 5 {
			t.Errorf("n = %d", n)
		}
		return 0
	})
}

func TestShape(t *testing.T) {
	sine := NewSineTable(4)
	tests := []struct{ x, want float64 }{
		{0, 0}, {0.125, 0.5}, {0.25, 1}, {0.5, 0}, {0.75, -1}, {0.875, -0.5},
		{1, 0},    // upper end
		{2, 0},    // above 1 is the end of the table
		{-0.5, 0}, // below 0 is the start of the table
	}
	for _, tt := range tests {
		if got := sine.Shape(tt.x); math.Abs(got-tt.want) > 1e-12 {
			t.Errorf("Shape(%v) = %v want %v", tt.x, got, tt.want)
		}
	}
	saw := NewSawtoothTable(4)
	if got := saw.Shape(0.875); math.Abs(got+0.25) > 1e-12 {
		t.Errorf("saw wraps down to the first value at 1: got %v", got)
	}
	if got := saw.Shape(1); got != -1 {
		t.Errorf("saw at 1: %v", got)
	}
	if got := (Table{0, 10}).Shape(0.3); math.Abs(got-3) > 1e-12 {
		t.Errorf("got %v", got)
	}
}

func TestDegenerateTables(t *testing.T) {
	for _, mode := range []FillMode{First, Last, Custom} {
		table := NewWithFunc(0, mode, func(i, n int) float64 { return 1 })
		if len(table) != 1 {
			t.Fatalf("mode %d: %v", mode, table)
		}
		if got := table.Shape(0.5); got != 0 {
			t.Fatalf("mode %d: got %v", mode, got)
		}
	}
	if got := (Table{4}).Shape(0.7); got != 4 {
		t.Errorf("single value table: %v", got)
	}
	if got := (Table{}).Shape(0.7); got != 0 {
		t.Errorf("empty table: %v", got)
	}
}
