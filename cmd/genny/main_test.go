package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/almerlucke/genny/midi"
)

const okSpec = `{"seed": 2, "tracks": [{"name": "d", "channel": 9, "source": {"kind": "euclid", "params": {"steps": 8, "pulses": 3}}}]}`

func runCLI(stdin string, args ...string) (code int, stdout, stderr string) {
	var out, errb bytes.Buffer
	code = run(args, strings.NewReader(stdin), &out, &errb)
	return code, out.String(), errb.String()
}

func writeSpec(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "piece.json")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRender(t *testing.T) {
	spec := writeSpec(t, okSpec)
	out := filepath.Join(t.TempDir(), "o.mid")
	for _, args := range [][]string{{"render", spec, "-o", out}, {"render", "-o", out, spec}} {
		os.Remove(out)
		code, stdout, stderr := runCLI("", args...)
		if code != 0 || !strings.Contains(stdout, "wrote") {
			t.Fatalf("%v: code %d %q %q", args, code, stdout, stderr)
		}
		f, err := midi.ReadFile(out)
		if err != nil || len(f.Tracks) != 1 || len(f.Tracks[0].Events) == 0 {
			t.Fatalf("read back: %v %+v", err, f)
		}
	}
}

func TestRenderDefaultOutputAndSeed(t *testing.T) {
	spec := writeSpec(t, okSpec)
	code, stdout, _ := runCLI("", "render", spec, "-seed", "9")
	if code != 0 || !strings.Contains(stdout, "seed 9") {
		t.Fatalf("%d %q", code, stdout)
	}
	if _, err := os.Stat(strings.TrimSuffix(spec, ".json") + ".mid"); err != nil {
		t.Error(err)
	}
}

func TestRenderStdin(t *testing.T) {
	code, stdout, stderr := runCLI(okSpec, "validate", "-")
	if code != 0 || !strings.Contains(stdout, "ok") {
		t.Fatalf("%d %q %q", code, stdout, stderr)
	}
}

func TestValidateErrors(t *testing.T) {
	spec := writeSpec(t, `{"tracks": [{"source": {"kind": "euclid", "params": {"steps": 8, "pulses": 99}}}]}`)
	code, _, stderr := runCLI("", "validate", spec)
	if code != 1 || !strings.Contains(stderr, "tracks[0].source.params.pulses: must be in 0..8, got 99") {
		t.Errorf("%d %q", code, stderr)
	}
}

func TestUsageErrors(t *testing.T) {
	spec := writeSpec(t, okSpec)
	cases := [][]string{
		{}, {"bogus"}, {"render"}, {"render", spec, spec}, {"render", spec, "-seed", "x"}, {"render", spec, "-seed", "-3"},
		{"render", spec, "-nope"}, {"list", "a", "b"},
	}
	for _, args := range cases {
		if code, _, _ := runCLI("", args...); code != 2 {
			t.Errorf("%v: code %d, want 2", args, code)
		}
	}
	if code, _, stderr := runCLI("", "render", "/nonexistent/x.json"); code != 1 || stderr == "" {
		t.Errorf("missing file: %d %q", code, stderr)
	}
	if code, _, stderr := runCLI("", "render", writeSpec(t, "{"), "-o", filepath.Join(t.TempDir(), "x.mid")); code != 1 || !strings.Contains(stderr, "line 1") {
		t.Errorf("bad json: %d %q", code, stderr)
	}
}

func TestList(t *testing.T) {
	code, stdout, _ := runCLI("", "list")
	if code != 0 {
		t.Fatal(code)
	}
	for _, want := range []string{"melody:", "euclid:", "sieve:", "concat:", "transpose:", "generators", "combinators", "minor_pentatonic"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("list lacks %q", want)
		}
	}
	code, stdout, _ = runCLI("", "list", "euclid")
	if code != 0 || !strings.Contains(stdout, "pulses") || strings.Contains(stdout, "melody:") {
		t.Errorf("list euclid: %d %q", code, stdout)
	}
	if code, _, _ := runCLI("", "list", "nope"); code != 1 {
		t.Errorf("unknown kind: %d", code)
	}
}
