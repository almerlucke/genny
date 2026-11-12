package spec

import (
	"os"
	"strings"
	"testing"
)

func TestReferenceCoversEveryKind(t *testing.T) {
	ref := Reference()
	for _, k := range Kinds() {
		if !strings.Contains(ref, "### "+k.Name+"\n") {
			t.Errorf("kind %s has no section", k.Name)
		}
		for _, p := range k.Params {
			if !strings.Contains(ref, "| `"+p.Name+"` |") {
				t.Errorf("%s: parameter %s is missing", k.Name, p.Name)
			}
		}
	}
}

// TestReferenceFileIsUpToDate fails when the kinds changed and docs/spec-reference.md was not regenerated
func TestReferenceFileIsUpToDate(t *testing.T) {
	data, err := os.ReadFile("../docs/spec-reference.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != Reference() {
		t.Error("docs/spec-reference.md is out of date: run `go run ./cmd/genny doc > docs/spec-reference.md`")
	}
}
