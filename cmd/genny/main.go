// Command genny renders a piece described in JSON to a MIDI file.
//
//	genny render spec.json -o out.mid [-seed N]
//	genny validate spec.json [-seed N]
//	genny list [kind]
//	genny doc
//
// See README.md in this directory for the format of a spec.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/almerlucke/genny/midi"
	"github.com/almerlucke/genny/spec"
)

const usage = `usage:
  genny render spec.json [-o out.mid] [-seed N]   render a spec to a MIDI file
  genny validate spec.json [-seed N]              check a spec without writing anything
  genny list [kind]                               list the kinds of nodes and their parameters
  genny doc                                       print the reference of the spec format as Markdown

spec.json may be - to read standard input.
`

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "render":
		return render(args[1:], stdin, stdout, stderr, true)
	case "validate":
		return render(args[1:], stdin, stdout, stderr, false)
	case "list":
		return list(args[1:], stdout, stderr)
	case "doc":
		fmt.Fprint(stdout, spec.Reference())
		return 0
	case "help", "-h", "-help", "--help":
		fmt.Fprint(stdout, usage)
		return 0
	}
	fmt.Fprintf(stderr, "genny: unknown command %q\n\n%s", args[0], usage)
	return 2
}

func render(args []string, stdin io.Reader, stdout, stderr io.Writer, write bool) int {
	name := "validate"
	if write {
		name = "render"
	}
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("o", "", "output MIDI file (default: the spec name with .mid)")
	seedFlag := fs.String("seed", "", "seed of the random stream, overrides the seed in the spec")
	// flags may come before or after the spec file
	var files []string
	rest := args
	for {
		if err := fs.Parse(rest); err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			return 2
		}
		if fs.NArg() == 0 {
			break
		}
		files = append(files, fs.Arg(0))
		rest = fs.Args()[1:]
	}
	if len(files) != 1 {
		fmt.Fprintf(stderr, "genny %s: expected exactly one spec file\n\n%s", name, usage)
		return 2
	}
	var seed *uint64
	if *seedFlag != "" {
		v, err := strconv.ParseUint(*seedFlag, 10, 64)
		if err != nil {
			fmt.Fprintf(stderr, "genny %s: -seed must be a non-negative integer, got %q\n", name, *seedFlag)
			return 2
		}
		seed = &v
	}

	path := files[0]
	var r io.Reader = stdin
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(stderr, "genny: %v\n", err)
			return 1
		}
		defer f.Close()
		r = f
	}
	s, err := spec.Parse(r)
	if err != nil {
		fmt.Fprintf(stderr, "genny: %s: %v\n", path, err)
		return 1
	}
	res, err := s.Render(seed)
	if err != nil {
		fmt.Fprintf(stderr, "genny: %s: %v\n", path, err)
		return 1
	}
	summary := fmt.Sprintf("%d tracks, %d notes, %.0f steps (%.1f s at %g bpm), seed %d",
		len(res.File.Tracks), res.Notes, res.Steps, res.Seconds(), res.File.Tempo, res.Seed)
	if !write {
		fmt.Fprintf(stdout, "%s: ok, %s\n", path, summary)
		return 0
	}
	dest := *out
	if dest == "" {
		if path == "-" {
			dest = "out.mid"
		} else {
			dest = strings.TrimSuffix(path, filepath.Ext(path)) + ".mid"
		}
	}
	if err := midi.WriteFile(dest, res.File); err != nil {
		fmt.Fprintf(stderr, "genny: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "wrote %s: %s\n", dest, summary)
	return 0
}

func list(args []string, stdout, stderr io.Writer) int {
	if len(args) > 1 {
		fmt.Fprintf(stderr, "genny list: expected at most one kind\n")
		return 2
	}
	if len(args) == 1 {
		k, ok := spec.Lookup(args[0])
		if !ok {
			fmt.Fprintf(stderr, "genny list: unknown kind %q\n", args[0])
			return 1
		}
		printKind(stdout, k)
		return 0
	}
	fmt.Fprintln(stdout, "A node is {\"kind\": NAME, \"params\": {...}}. Steps are the unit of time (sixteenths by default).")
	fmt.Fprintln(stdout, "Pitches are numbers or names (C4 = 60). Scales are names or arrays of semitone steps:")
	fmt.Fprintf(stdout, "  %s\n", strings.Join(spec.ScaleNames(), ", "))
	cat := ""
	for _, k := range spec.Kinds() {
		if k.Category != cat {
			cat = k.Category
			fmt.Fprintf(stdout, "\n%ss\n", cat)
		}
		printKind(stdout, k)
	}
	return 0
}

func printKind(w io.Writer, k spec.Kind) {
	fmt.Fprintf(w, "\n%s: %s\n", k.Name, k.Doc)
	for _, p := range k.Params {
		def := p.Default
		if def == "" {
			def = "required"
		}
		fmt.Fprintf(w, "    %-14s %-16s [%s] %s\n", p.Name, p.Type, def, p.Doc)
	}
}
