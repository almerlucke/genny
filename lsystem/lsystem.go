// Package lsystem rewrites a string with rules (a Lindenmayer system) and plays the result as a
// sequence, which gives self-similar phrases
package lsystem

import (
	"strings"

	"github.com/almerlucke/genny/sequence"
)

// MaxLen limits the length of the expanded string, longer results are cut off
const MaxLen = 1 << 16

// Expand rewrites axiom iterations times, every symbol with a rule is replaced by the rule, other
// symbols stay as they are
func Expand(axiom string, rules map[rune]string, iterations int) string {
	cur := axiom
	for range iterations {
		var b strings.Builder
		for _, r := range cur {
			if rule, ok := rules[r]; ok {
				b.WriteString(rule)
			} else {
				b.WriteRune(r)
			}
			if b.Len() >= MaxLen {
				break
			}
		}
		cur = b.String()
		if len(cur) >= MaxLen {
			cur = cur[:MaxLen]
			break
		}
	}
	return cur
}

func values[T any](axiom string, rules map[rune]string, iterations int, mapping map[rune]T) []T {
	var vals []T
	for _, r := range Expand(axiom, rules, iterations) {
		if v, ok := mapping[r]; ok {
			vals = append(vals, v)
		}
	}
	if len(vals) == 0 {
		var zero T
		vals = append(vals, zero)
	}
	return vals
}

// New expands the axiom and plays the symbols once, each symbol is turned into a value with mapping.
// Symbols without a mapping (for instance structure symbols) are skipped. If nothing is left the
// sequence is a single zero value
func New[T any](axiom string, rules map[rune]string, iterations int, mapping map[rune]T) *sequence.Sequence[T] {
	return sequence.New(values(axiom, rules, iterations, mapping)...)
}

// NewLoop is like New but loops
func NewLoop[T any](axiom string, rules map[rune]string, iterations int, mapping map[rune]T) *sequence.Sequence[T] {
	return sequence.NewLoop(values(axiom, rules, iterations, mapping)...)
}
