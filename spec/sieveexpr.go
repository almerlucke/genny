package spec

import (
	"fmt"
	"strconv"

	"github.com/almerlucke/genny/sieve"
)

// ParseSieve parses a sieve expression over the range [lo, hi]. The atoms are residue classes
// "mod@start" (3@1 is 1, 4, 7, ...); they combine with "&" (intersection), "|" (union), "-" (difference),
// "!" (complement within the range) and parentheses. "&" binds tighter than "|" and "-", which
// associate to the left: "(3@0|5@1)&!2@0"
func ParseSieve(expr string, lo, hi int) (s sieve.Sieve, err error) {
	p := &sieveParser{src: expr, lo: lo, hi: hi}
	p.skip()
	if p.pos >= len(p.src) {
		return s, fmt.Errorf("empty expression")
	}
	s, err = p.union()
	if err != nil {
		return s, err
	}
	if p.pos < len(p.src) {
		return s, p.errorf("unexpected %q", p.src[p.pos])
	}
	return s, nil
}

type sieveParser struct {
	src    string
	pos    int
	lo, hi int
	depth  int
}

func (p *sieveParser) errorf(format string, args ...any) error {
	return fmt.Errorf("position %d: %s", p.pos+1, fmt.Sprintf(format, args...))
}

func (p *sieveParser) skip() {
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}
}

func (p *sieveParser) peek() byte {
	p.skip()
	if p.pos < len(p.src) {
		return p.src[p.pos]
	}
	return 0
}

func (p *sieveParser) union() (sieve.Sieve, error) {
	left, err := p.inter()
	if err != nil {
		return left, err
	}
	for {
		switch p.peek() {
		case '|':
			p.pos++
			right, err := p.inter()
			if err != nil {
				return left, err
			}
			left = left.Union(right)
		case '-':
			p.pos++
			right, err := p.inter()
			if err != nil {
				return left, err
			}
			left = left.Difference(right)
		default:
			return left, nil
		}
	}
}

func (p *sieveParser) inter() (sieve.Sieve, error) {
	left, err := p.unary()
	if err != nil {
		return left, err
	}
	for p.peek() == '&' {
		p.pos++
		right, err := p.unary()
		if err != nil {
			return left, err
		}
		left = left.Intersection(right)
	}
	return left, nil
}

func (p *sieveParser) unary() (sieve.Sieve, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > 64 {
		return sieve.Sieve{}, p.errorf("expression nested too deeply")
	}
	switch c := p.peek(); {
	case c == '!':
		p.pos++
		s, err := p.unary()
		return s.Complement(), err
	case c == '(':
		p.pos++
		s, err := p.union()
		if err != nil {
			return s, err
		}
		if p.peek() != ')' {
			return s, p.errorf("expected ')'")
		}
		p.pos++
		return s, nil
	case c >= '0' && c <= '9':
		mod, err := p.number()
		if err != nil {
			return sieve.Sieve{}, err
		}
		if mod < 1 {
			return sieve.Sieve{}, p.errorf("modulus must be at least 1, got %d", mod)
		}
		if p.peek() != '@' {
			return sieve.Sieve{}, p.errorf("expected '@' after the modulus, as in %d@0", mod)
		}
		p.pos++
		p.skip()
		start, err := p.number()
		if err != nil {
			return sieve.Sieve{}, err
		}
		return sieve.New(mod, start, p.lo, p.hi), nil
	case c == 0:
		return sieve.Sieve{}, p.errorf("unexpected end of expression")
	default:
		return sieve.Sieve{}, p.errorf("unexpected %q", c)
	}
}

func (p *sieveParser) number() (int, error) {
	p.skip()
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] >= '0' && p.src[p.pos] <= '9' {
		p.pos++
	}
	if start == p.pos {
		return 0, p.errorf("expected a number")
	}
	n, err := strconv.Atoi(p.src[start:p.pos])
	if err != nil || n > 1<<30 {
		return 0, p.errorf("number %q is too large", p.src[start:p.pos])
	}
	return n, nil
}
