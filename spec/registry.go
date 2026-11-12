package spec

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/almerlucke/genny/note"
)

// Line is the value of a node: notes and the length of the line in steps (the span a repeat or concat
// uses, which can be longer than the last note ends)
type Line struct {
	Events []note.Event
	Length float64
}

// Param documents one parameter of a kind, for `genny list`
type Param struct {
	Name string
	// Type is a short description such as "int", "pitch", "scale", "node"
	Type string
	// Default is the default value as text, empty if the parameter is required
	Default string
	Doc     string
}

// Kind is one kind of node. Adding a kind is one Register call, see kinds_*.go
type Kind struct {
	Name string
	// Category is "generator" (makes notes) or "combinator" (works on child nodes)
	Category string
	Doc      string
	Params   []Param
	// Build builds the node. params is the raw "params" object (possibly empty); path is the JSON path of
	// the node. Decode it with Builder.Decode and report problems with Fail at path+".params.<field>"
	Build func(b *Builder, path string, params json.RawMessage) (Line, error)
}

var kinds = map[string]Kind{}

// Register adds a kind to the table. It panics on a duplicate or incomplete kind (a programming error)
func Register(k Kind) {
	if k.Name == "" || k.Build == nil || k.Doc == "" {
		panic("spec: a kind needs a name, a doc and a builder")
	}
	if _, dup := kinds[k.Name]; dup {
		panic("spec: duplicate kind " + k.Name)
	}
	kinds[k.Name] = k
}

// Kinds returns all kinds, generators first, then combinators, each sorted by name
func Kinds() []Kind {
	out := make([]Kind, 0, len(kinds))
	for _, k := range kinds {
		out = append(out, k)
	}
	slices.SortFunc(out, func(a, b Kind) int {
		if a.Category != b.Category {
			return strings.Compare(b.Category, a.Category) // generator before combinator
		}
		return strings.Compare(a.Name, b.Name)
	})
	return out
}

// Lookup finds a kind by name
func Lookup(name string) (Kind, bool) {
	k, ok := kinds[name]
	return k, ok
}

// Builder builds nodes. It counts the work so a spec cannot ask for unbounded output
type Builder struct {
	depth int
	notes int
}

// nodeDoc is the shape of every node
type nodeDoc struct {
	Kind   string          `json:"kind"`
	Params json.RawMessage `json:"params"`
}

// Child builds the node in raw, found at path. It recovers panics of the library constructors and turns
// them into errors, so a bad spec cannot crash the tool
func (b *Builder) Child(path string, raw json.RawMessage) (line Line, err error) {
	if len(raw) == 0 || string(raw) == "null" {
		return Line{}, Fail(path, "is required")
	}
	var n nodeDoc
	if err := decodeStrict(path, raw, &n); err != nil {
		return Line{}, err
	}
	if n.Kind == "" {
		return Line{}, Fail(path+".kind", "is required")
	}
	k, ok := kinds[n.Kind]
	if !ok {
		return Line{}, Fail(path+".kind", "unknown kind %q%s", n.Kind, suggest(n.Kind))
	}
	if b.depth >= maxDepth {
		return Line{}, Fail(path, "nested deeper than %d levels", maxDepth)
	}
	params := n.Params
	if len(params) == 0 || string(params) == "null" {
		params = json.RawMessage("{}")
	}
	b.depth++
	defer func() {
		b.depth--
		if r := recover(); r != nil {
			line, err = Line{}, Fail(path, "kind %q failed: %v", n.Kind, r)
		}
	}()
	line, err = k.Build(b, path, params)
	if err != nil {
		return Line{}, err
	}
	if !(line.Length >= 0 && line.Length <= maxSteps) {
		return Line{}, Fail(path, "the line is %v steps long, at most %d are allowed", line.Length, maxSteps)
	}
	if err := b.count(path, len(line.Events)); err != nil {
		return Line{}, err
	}
	return line, nil
}

// count charges notes to the budget of the render. A node is charged for what it holds, so nested
// combinators are charged repeatedly: that only makes the limit stricter
func (b *Builder) count(path string, n int) error {
	if n > maxNodeNotes*5 {
		return Fail(path, "makes %d notes, too many", n)
	}
	b.notes += n
	if b.notes > maxTotalNotes*4 {
		return Fail(path, "the spec makes too many notes in total (limit about %d)", maxTotalNotes)
	}
	return nil
}

// Decode decodes the params of a node into v, which should hold the defaults: absent fields keep them.
// Unknown fields and wrong types are errors quoting path+".params"
func (b *Builder) Decode(path string, params json.RawMessage, v any) error {
	return decodeStrict(path+".params", params, v)
}

// suggest names the closest kind, if one is close
func suggest(name string) string {
	best, bestD := "", 3
	for k := range kinds {
		if d := editDistance(name, k); d < bestD {
			best, bestD = k, d
		}
	}
	if best != "" {
		return fmt.Sprintf(" (did you mean %q?)", best)
	}
	return " (see `genny list`)"
}

func editDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}
