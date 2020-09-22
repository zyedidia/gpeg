package pattern

import (
	"fmt"

	"github.com/zyedidia/gpeg/isa"
)

// A Pattern is a set of instructions that can be used to
// match an input Reader.
type Pattern []isa.Insn

// String returns the string representation of a pattern
func (p Pattern) String() string {
	s := ""
	for i, instr := range p {
		s += fmt.Sprintf("%2d: %v\n", i, instr)
	}
	return s
}

// Literal matches a given string literal.
func Literal(s string) Pattern {
	code := make(Pattern, len(s))
	for i := 0; i < len(s); i++ {
		code[i] = isa.Char{Byte: s[i]}
	}
	return code
}

// Set matches any character in the given set.
func Set(chars isa.Charset) Pattern {
	return Pattern{isa.Set{Chars: chars}}
}

// Any consumes n characters, and only fails if there
// aren't enough input characters left.
func Any(n uint8) Pattern {
	return Pattern{isa.Any{N: n}}
}

// Concat concatenates two patterns: `p1 p2`.
func Concat(p1, p2 Pattern) Pattern {
	code := make(Pattern, 0, len(p1)+len(p2))
	code = append(code, p1...)
	code = append(code, p2...)
	return code
}

// Or returns the ordered choice between two patterns: `p1 / p2`.
func Or(p1, p2 Pattern) Pattern {
	// optimization: if p1 and p2 are charsets, return the union
	if len(p1) == 1 && len(p2) == 1 {
		s1, ok1 := p1[0].(isa.Set)
		s2, ok2 := p2[0].(isa.Set)
		if ok1 && ok2 {
			return Pattern{
				isa.Set{Chars: s1.Chars.Add(s2.Chars)},
			}
		}
	}

	code := make(Pattern, 0, len(p1)+len(p2)+1+1+1+1)
	L1 := isa.NewLabel()
	L2 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, p1...)
	code = append(code, isa.Commit{Lbl: L2})
	code = append(code, L1)
	code = append(code, p2...)
	code = append(code, L2)
	return code
}

// Star returns the Kleene star repetition of a pattern: `p*`.
// This matches zero or more occurrences of p.
func Star(p Pattern) Pattern {
	// optimization: repeating a charset uses the dedicated instruction 'span'
	if len(p) == 1 {
		s, ok := p[0].(isa.Set)
		if ok {
			return Pattern{
				isa.Span{Chars: s.Chars},
			}
		}
	}

	code := make(Pattern, 0, len(p)+4)
	L1 := isa.NewLabel()
	L2 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L2})
	code = append(code, L1)
	code = append(code, p...)
	code = append(code, isa.PartialCommit{Lbl: L1})
	code = append(code, L2)
	return code
}

// Plus returns the Kleene plus repetition of a pattern: `p+`.
// This matches one or more occurrences of p.
func Plus(p Pattern) Pattern {
	starp := Star(p)
	code := make(Pattern, 0, len(p)+len(starp))
	code = append(code, p...)
	code = append(code, starp...)
	return code
}

// Optional matches at most 1 occurrence of p: `p?`.
func Optional(p Pattern) Pattern {
	return Or(p, Pattern{})
}

// Not returns the not predicate applied to a pattern: `!p`.
// The not predicate succeeds if matching `p` at the current position
// fails, and does not consume any input.
func Not(p Pattern) Pattern {
	code := make(Pattern, 0, len(p)+3)
	L1 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, p...)
	code = append(code, isa.FailTwice{})
	code = append(code, L1)
	return code
}

// And returns the and predicate applied to a pattern: `&p`.
// The and predicate succeeds if matching `p` at the current position
// succeeds and does not consume any input.
// This is equivalent to `!!p`.
func And(p Pattern) Pattern {
	code := make(Pattern, 0, len(p)+7)
	L1 := isa.NewLabel()
	L2 := isa.NewLabel()
	L3 := isa.NewLabel()
	code = append(code, isa.Choice{Lbl: L1})
	code = append(code, isa.Choice{Lbl: L2})
	code = append(code, p...)
	code = append(code, L2)
	code = append(code, isa.Commit{Lbl: L3})
	code = append(code, L3)
	code = append(code, isa.Fail{})
	code = append(code, L1)
	return code
}
